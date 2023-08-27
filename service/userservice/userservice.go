package userservice

import (
	"noteservice/model"
)

type UserStore interface {
	CreateUser(user model.User) error
	GetUser(username string) (model.User, error)
}

type UserService interface {
	CreateUser(user model.User) error
	GetUser(username string) (model.User, error)
}

type userService struct {
	store UserStore
}

func NewUserService(store UserStore) UserService {
	return userService{
		store: store,
	}
}

func (u userService) CreateUser(user model.User) error {
	return u.store.CreateUser(user)
}

func (u userService) GetUser(username string) (model.User, error) {
	return u.store.GetUser(username)
}