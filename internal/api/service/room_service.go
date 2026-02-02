package service

import (
	"context"
	"fmt"

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
	JoinRoom(roomID uint, userID uint) error // ⭐ 추가: 채팅방 입장
	AddMember(roomID uint, userID, adderID uint) error
	RemoveMember(roomID uint, userID, removerID uint) error
	LeaveRoom(roomID uint, userID uint) error
}

type roomService struct {
	roomRepo  repository.RoomRepository
	rdb       *redis.Client
	publisher *pubsub.Publisher
}

func NewRoomService(
	roomRepo repository.RoomRepository,
	rdb *redis.Client,
	publisher *pubsub.Publisher,
) RoomService {
	return &roomService{roomRepo: roomRepo, rdb: rdb, publisher: publisher}
}

// CreateRoom 채팅방 생성
func (s *roomService) CreateRoom(req *dto.CreateRoomRequest, creatorID uint) (*model.Room, error) {
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

	// 5. ⭐ Redis Pub/Sub 이벤트 발행 (모든 WS 서버에 알림)
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

// JoinRoom - 채팅방 입장 (입장 알림 발송)
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

	// 2. ⭐ Redis Pub/Sub으로 입장 알림 발행
	event := map[string]interface{}{
		"type":     "user_joined",
		"room_id":  roomID,
		"user_id":  userID,
		"nickname": nickname,
	}
	if err := s.publisher.PublishRoomEvent(event); err != nil {
		fmt.Printf("Failed to publish user_joined event: %v\n", err)
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

func (s *roomService) LeaveRoom(roomID uint, userID uint) error {
	// 1. DB 업데이트
	if err := s.roomRepo.RemoveMember(roomID, userID); err != nil {
		return errors.WrapDatabase(err, "Failed to leave room")
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
