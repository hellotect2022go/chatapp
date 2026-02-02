package service

import (
	"encoding/json"

	"github.com/hellotect2022go/chatapp/internal/api/dto"
	"github.com/hellotect2022go/chatapp/internal/api/repository"
	"github.com/hellotect2022go/chatapp/internal/shared/errors"
	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"github.com/hellotect2022go/chatapp/internal/shared/pubsub"
)

type MessageService interface {
	SendMessage(roomID, userID uint, req *dto.SendMessageRequest) (*model.Message, error)
	SaveMessage(message *model.Message) error
	PublishMessage(message *model.Message) error // ⭐ 추가: Redis Pub/Sub 발행 전용
	GetMessages(roomID uint, userID uint, page, pageSize int) (*dto.MessageListResponse, error)
}

type messageService struct {
	messageRepository repository.MessageRepository
	roomRepository    repository.RoomRepository
	userRepository    repository.UserRepository
	publisher         *pubsub.Publisher
}

func NewMessageService(
	messageRepository repository.MessageRepository,
	roomRepository repository.RoomRepository,
	userRepository repository.UserRepository,
	publisher *pubsub.Publisher,
) MessageService {
	return &messageService{
		messageRepository: messageRepository,
		roomRepository:    roomRepository,
		userRepository:    userRepository,
		publisher:         publisher,
	}
}

// ✅ 메시지 전송
func (s *messageService) SendMessage(roomID uint, userID uint, req *dto.SendMessageRequest) (*model.Message, error) {
	// 권한 확인: 해당 방의 멤버인지
	members, err := s.roomRepository.GetMembers(roomID)
	if err != nil {
		return nil, errors.WrapDatabase(err, "Failed to get room members")
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

	// 메시지 저장
	message := &model.Message{
		RoomID:           roomID,
		UserID:           userID,
		Content:          req.Content,
		Type:             req.Type,
		MessageEventType: "chat", // ⭐ 채팅 메시지
	}

	if err := s.messageRepository.Create(message); err != nil {
		return nil, errors.WrapDatabase(err, "Failed to save message")
	}

	// 사용자 정보 로드 (Nickname 포함)
	user, err := s.userRepository.FindByID(userID)
	if err != nil {
		return nil, errors.WrapDatabase(err, "Failed to load user")
	}

	// Redis Pub/Sub로 발행 (실시간 알림)
	wsMsg := &model.WSMessage{
		Type:        model.MessageTypeChat,
		MessageID:   message.ID,
		RoomID:      message.RoomID,
		UserID:      message.UserID,
		Nickname:    user.Nickname,
		Content:     message.Content,
		ContentType: message.Type, // ⭐ text, image, file
		Timestamp:   message.CreatedAt,
	}

	msgBytes, _ := json.Marshal(wsMsg)
	s.publisher.Publish(roomID, msgBytes)

	return message, nil
}

func (s *messageService) SaveMessage(message *model.Message) error {
	if err := s.messageRepository.Create(message); err != nil {
		return errors.WrapDatabase(err, "Failed to save message")
	}
	return nil
}

// ⭐ Redis Pub/Sub으로 메시지 발행 (파일 정보 포함)
func (s *messageService) PublishMessage(message *model.Message) error {
	// 사용자 정보 로드 (Nickname 포함)
	user, err := s.userRepository.FindByID(message.UserID)
	if err != nil {
		return errors.WrapDatabase(err, "Failed to load user")
	}

	// Redis Pub/Sub로 발행 (실시간 알림)
	wsMsg := &model.WSMessage{
		Type:        model.MessageTypeChat,
		MessageID:   message.ID,
		RoomID:      message.RoomID,
		UserID:      message.UserID,
		Nickname:    user.Nickname,
		Content:     message.Content,
		ContentType: message.Type, // ⭐ text, image, file
		Timestamp:   message.CreatedAt,
		Files:       message.Files,
	}

	msgBytes, err := json.Marshal(wsMsg)
	if err != nil {
		return errors.WrapInternal(err, "Failed to marshal message")
	}

	s.publisher.Publish(message.RoomID, msgBytes)
	return nil
}

func (s *messageService) GetMessages(roomID uint, userID uint, page, pageSize int) (*dto.MessageListResponse, error) {
	// 권한 확인: 해당 방의 멤버인지
	members, err := s.roomRepository.GetMembers(roomID)
	if err != nil {
		return nil, errors.WrapDatabase(err, "Failed to get room members")
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

	// 페이지네이션
	offset := (page - 1) * pageSize

	messages, err := s.messageRepository.FindByRoomID(roomID, pageSize, offset)
	if err != nil {
		return nil, errors.WrapDatabase(err, "Failed to load messages")
	}

	total, err := s.messageRepository.CountByRoomID(roomID)
	if err != nil {
		return nil, errors.WrapDatabase(err, "Failed to count messages")
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return &dto.MessageListResponse{
		Messages:   messages,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
