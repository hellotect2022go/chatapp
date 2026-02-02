package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hellotect2022go/chatapp/internal/api/dto"
	"github.com/hellotect2022go/chatapp/internal/api/repository"
	"github.com/hellotect2022go/chatapp/internal/shared/errors"
	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"github.com/hellotect2022go/chatapp/internal/shared/pubsub"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type RoomService interface {
	CreateRoom(req *dto.CreateRoomRequest, creatorID uint) (*model.Room, error)
	GetMyRooms(userID uint) ([]*model.Room, error)
	GetRoom(roomID uint, userID uint) (*model.Room, error)
	JoinRoom(roomID uint, userID uint) error            // ⭐ WebSocket 알림만 (메시지 저장 안 함)
	JoinRoomWithMessage(roomID uint, userID uint) error // ⭐ 메시지 저장 + WebSocket 알림
	AddMember(roomID uint, userID, adderID uint) error
	RemoveMember(roomID uint, userID, removerID uint) error
	LeaveRoom(roomID uint, userID uint) error
}

type roomService struct {
	roomRepo    repository.RoomRepository
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
	rdb         *redis.Client
	publisher   *pubsub.Publisher
}

func NewRoomService(
	roomRepo repository.RoomRepository,
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
	rdb *redis.Client,
	publisher *pubsub.Publisher,
) RoomService {
	return &roomService{
		roomRepo:    roomRepo,
		messageRepo: messageRepo,
		userRepo:    userRepo,
		rdb:         rdb,
		publisher:   publisher,
	}
}

// CreateRoom 채팅방 생성
func (s *roomService) CreateRoom(req *dto.CreateRoomRequest, creatorID uint) (*model.Room, error) {
	// ⭐ Direct 채팅방인 경우 중복 확인
	if req.Type == "direct" {
		// Direct는 정확히 1명의 상대방과만 가능
		if len(req.Members) != 1 {
			return nil, errors.BadRequest("Direct chat requires exactly one other member")
		}

		otherUserID := req.Members[0]

		// 본인과의 채팅 방지
		if otherUserID == creatorID {
			return nil, errors.BadRequest("Cannot create direct chat with yourself")
		}

		// 기존 Direct 방 확인
		existingRoom, err := s.roomRepo.FindDirectRoomByUsers(creatorID, otherUserID)
		if err == nil && existingRoom != nil {
			// 기존 방이 있으면 해당 방 반환
			return existingRoom, nil
		}
		// gorm.ErrRecordNotFound인 경우 새로 생성 진행
	}

	// 1. Room 생성 (DB)
	room := &model.Room{
		Name:      req.Name,
		Type:      model.RoomType(req.Type),
		CreatedBy: creatorID,
	}

	if err := s.roomRepo.Create(room); err != nil {
		return nil, errors.WrapDatabase(err, "Failed to create room")
	}

	// 2. 생성자를 멤버로 추가 (DB)
	if err := s.roomRepo.AddMember(room.ID, creatorID); err != nil {
		return nil, errors.WrapDatabase(err, "Failed to add creator as member")
	}

	// 3. 초대된 사용자들 추가 (DB)
	allMembers := []uint{creatorID}
	for _, userID := range req.Members {
		if userID != creatorID {
			if err := s.roomRepo.AddMember(room.ID, userID); err != nil {
				// 일부 멤버 추가 실패는 로그만 남기고 계속
				fmt.Printf("Warning: Failed to add member %d: %v\n", userID, err)
			} else {
				allMembers = append(allMembers, userID)
			}
		}
	}

	// 4. ⭐ Redis 업데이트 (user:X:rooms, room:X:members)
	ctx := context.Background()
	for _, userID := range allMembers {
		userRoomsKey := fmt.Sprintf("user:%d:rooms", userID)
		s.rdb.SAdd(ctx, userRoomsKey, room.ID)
	}
	roomMembersKey := fmt.Sprintf("room:%d:members", room.ID)
	for _, userID := range allMembers {
		s.rdb.SAdd(ctx, roomMembersKey, userID)
	}

	// 5. ⭐ 모든 멤버의 입장 메시지 생성 (최초 입장)
	for _, userID := range allMembers {
		user, err := s.userRepo.FindByID(userID)
		if err != nil {
			fmt.Printf("Warning: Failed to load user %d for join message: %v\n", userID, err)
			continue
		}

		// 입장 메시지 DB 저장
		joinMessage := &model.Message{
			RoomID:           room.ID,
			UserID:           userID,
			Content:          fmt.Sprintf("%s님이 입장하셨습니다.", user.Nickname),
			Type:             "text",
			MessageEventType: model.MessageTypeJoin,
		}

		if err := s.messageRepo.Create(joinMessage); err != nil {
			fmt.Printf("Warning: Failed to create join message for user %d: %v\n", userID, err)
		}

		// WebSocket 실시간 알림
		wsMsg := &model.WSMessage{
			Type:      model.MessageTypeJoin,
			RoomID:    room.ID,
			UserID:    userID,
			Nickname:  user.Nickname,
			Content:   joinMessage.Content,
			Timestamp: time.Now(),
		}

		msgBytes, err := json.Marshal(wsMsg)
		if err == nil {
			s.publisher.Publish(room.ID, msgBytes)
		}
	}

	// 6. ⭐ Redis Pub/Sub 이벤트 발행 (모든 WS 서버에 알림)
	event := map[string]interface{}{
		"type":    "room_created",
		"room_id": room.ID,
		"members": allMembers,
	}
	if err := s.publisher.PublishRoomEvent(event); err != nil {
		// 이벤트 발행 실패는 로그만 남기고 계속 진행
		fmt.Printf("Failed to publish room_created event: %v\n", err)
	}

	return room, nil
}

// GetMyRooms 내가 참여한 채팅방 목록 조회
func (s *roomService) GetMyRooms(userID uint) ([]*model.Room, error) {
	rooms, err := s.roomRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.WrapDatabase(err, "Failed to load rooms")
	}
	return rooms, nil
}

// GetRoom 채팅방 조회
func (s *roomService) GetRoom(roomID uint, userID uint) (*model.Room, error) {
	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("Room")
		}
		return nil, errors.WrapDatabase(err, "Failed to find room")
	}

	// 권한 확인: 해당 방의 멤버인지
	members, err := s.roomRepo.GetMembers(roomID)
	if err != nil {
		return nil, errors.WrapDatabase(err, "Failed to get members")
	}

	isMember := false
	for _, member := range members {
		if member.ID == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, errors.Forbidden("You are not a member of this room")
	}

	return room, nil
}

// JoinRoom - WebSocket 실시간 알림만 (메시지 저장 안 함)
// 이미 멤버인 사용자가 방을 다시 선택할 때 사용
func (s *roomService) JoinRoom(roomID uint, userID uint) error {
	// 1. 권한 확인: 해당 방의 멤버인지
	members, err := s.roomRepo.GetMembers(roomID)
	if err != nil {
		return errors.WrapDatabase(err, "Failed to get members")
	}

	isMember := false
	var nickname string
	for _, member := range members {
		if member.ID == userID {
			isMember = true
			nickname = member.Nickname
			break
		}
	}
	if !isMember {
		return errors.Forbidden("You are not a member of this room")
	}

	// 2. ⭐ Redis Pub/Sub으로 입장 알림만 발행 (DB 저장 안 함)
	content := fmt.Sprintf("%s님이 입장하셨습니다.", nickname)
	wsMsg := &model.WSMessage{
		Type:      model.MessageTypeJoin,
		RoomID:    roomID,
		UserID:    userID,
		Nickname:  nickname,
		Content:   content,
		Timestamp: time.Now(),
	}

	msgBytes, err := json.Marshal(wsMsg)
	if err != nil {
		return errors.WrapInternal(err, "Failed to marshal join message")
	}

	if err := s.publisher.Publish(roomID, msgBytes); err != nil {
		return errors.WrapInternal(err, "Failed to publish join event")
	}

	return nil
}

// JoinRoomWithMessage - DB에 입장 메시지 저장 + WebSocket 알림
// 새로운 멤버가 방에 초대되어 처음 입장할 때 사용
func (s *roomService) JoinRoomWithMessage(roomID uint, userID uint) error {
	// 1. 권한 확인: 해당 방의 멤버인지
	members, err := s.roomRepo.GetMembers(roomID)
	if err != nil {
		return errors.WrapDatabase(err, "Failed to get members")
	}

	isMember := false
	var nickname string
	for _, member := range members {
		if member.ID == userID {
			isMember = true
			nickname = member.Nickname
			break
		}
	}
	if !isMember {
		return errors.Forbidden("You are not a member of this room")
	}

	// 2. ⭐ 입장 메시지 DB 저장
	joinMessage := &model.Message{
		RoomID:           roomID,
		UserID:           userID,
		Content:          fmt.Sprintf("%s님이 입장하셨습니다.", nickname),
		Type:             "text",
		MessageEventType: model.MessageTypeJoin,
	}

	if err := s.messageRepo.Create(joinMessage); err != nil {
		return errors.WrapDatabase(err, "Failed to save join message")
	}

	// 3. ⭐ Redis Pub/Sub으로 실시간 알림 발행
	wsMsg := &model.WSMessage{
		Type:      model.MessageTypeJoin,
		RoomID:    roomID,
		UserID:    userID,
		Nickname:  nickname,
		Content:   joinMessage.Content,
		Timestamp: time.Now(),
	}

	msgBytes, err := json.Marshal(wsMsg)
	if err != nil {
		return errors.WrapInternal(err, "Failed to marshal join message")
	}

	if err := s.publisher.Publish(roomID, msgBytes); err != nil {
		return errors.WrapInternal(err, "Failed to publish join event")
	}

	return nil
}

func (s *roomService) AddMember(roomID uint, userID, adderID uint) error {
	// 1. DB 업데이트
	if err := s.roomRepo.AddMember(roomID, userID); err != nil {
		return errors.WrapDatabase(err, "Failed to add member")
	}

	// 2. ⭐ Redis 업데이트
	ctx := context.Background()
	userRoomsKey := fmt.Sprintf("user:%d:rooms", userID)
	s.rdb.SAdd(ctx, userRoomsKey, roomID)

	roomMembersKey := fmt.Sprintf("room:%d:members", roomID)
	s.rdb.SAdd(ctx, roomMembersKey, userID)

	// 3. ⭐ Redis Pub/Sub 이벤트 발행
	event := map[string]interface{}{
		"type":    "member_added",
		"room_id": roomID,
		"user_id": userID,
	}
	if err := s.publisher.PublishRoomEvent(event); err != nil {
		fmt.Printf("Failed to publish member_added event: %v\n", err)
	}

	return nil
}

func (s *roomService) RemoveMember(roomID uint, userID, removerID uint) error {
	// 1. DB 업데이트
	if err := s.roomRepo.RemoveMember(roomID, userID); err != nil {
		return errors.WrapDatabase(err, "Failed to remove member")
	}

	// 2. ⭐ Redis 업데이트
	ctx := context.Background()
	userRoomsKey := fmt.Sprintf("user:%d:rooms", userID)
	s.rdb.SRem(ctx, userRoomsKey, roomID)

	roomMembersKey := fmt.Sprintf("room:%d:members", roomID)
	s.rdb.SRem(ctx, roomMembersKey, userID)

	// 3. ⭐ Redis Pub/Sub 이벤트 발행
	event := map[string]interface{}{
		"type":    "member_removed",
		"room_id": roomID,
		"user_id": userID,
	}
	if err := s.publisher.PublishRoomEvent(event); err != nil {
		fmt.Printf("Failed to publish member_removed event: %v\n", err)
	}

	return nil
}

// LeaveRoom - 채팅방 퇴장 (DB 저장 및 WebSocket 알림)
// LeaveRoom - 채팅방 퇴장 (DB에 퇴장 메시지 저장 + WebSocket 알림)
func (s *roomService) LeaveRoom(roomID uint, userID uint) error {
	// 1. 사용자 정보 조회 (nickname 필요)
	members, err := s.roomRepo.GetMembers(roomID)
	if err != nil {
		return errors.WrapDatabase(err, "Failed to get members")
	}

	var nickname string
	for _, member := range members {
		if member.ID == userID {
			nickname = member.Nickname
			break
		}
	}

	// 2. ⭐ 퇴장 메시지 DB 저장
	leaveMessage := &model.Message{
		RoomID:           roomID,
		UserID:           userID,
		Content:          fmt.Sprintf("%s님이 퇴장하셨습니다.", nickname),
		Type:             "text",
		MessageEventType: model.MessageTypeLeave,
	}

	if err := s.messageRepo.Create(leaveMessage); err != nil {
		return errors.WrapDatabase(err, "Failed to save leave message")
	}

	// 3. DB에서 멤버 제거
	if err := s.roomRepo.RemoveMember(roomID, userID); err != nil {
		return errors.WrapDatabase(err, "Failed to leave room")
	}

	// 4. ⭐ Redis 업데이트
	ctx := context.Background()
	userRoomsKey := fmt.Sprintf("user:%d:rooms", userID)
	s.rdb.SRem(ctx, userRoomsKey, roomID)

	roomMembersKey := fmt.Sprintf("room:%d:members", roomID)
	s.rdb.SRem(ctx, roomMembersKey, userID)

	// 5. ⭐ Redis Pub/Sub으로 퇴장 알림 발행 (WebSocket으로 실시간 전달)
	wsMsg := &model.WSMessage{
		Type:      model.MessageTypeLeave,
		RoomID:    roomID,
		UserID:    userID,
		Nickname:  nickname,
		Content:   leaveMessage.Content,
		Timestamp: time.Now(),
	}

	msgBytes, err := json.Marshal(wsMsg)
	if err != nil {
		return errors.WrapInternal(err, "Failed to marshal leave message")
	}

	if err := s.publisher.Publish(roomID, msgBytes); err != nil {
		return errors.WrapInternal(err, "Failed to publish leave event")
	}

	return nil
}
