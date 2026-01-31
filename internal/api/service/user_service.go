package service

type UserService interface {
	//GetByID(id uint)
}

type userService struct {
}

func NewUserService() UserService {
	return &userService{}
}

func (s *userService) GetByID(id uint) {
	//return nil
}
