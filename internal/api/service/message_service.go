package service

type MessageService interface {
}

type messageService struct {
}

func NewMessageService() MessageService {
	return &messageService{}
}
