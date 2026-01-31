package service

type RoomService interface {
}

type roomService struct {
}

func NewRoomService() RoomService {
	return &roomService{}
}
