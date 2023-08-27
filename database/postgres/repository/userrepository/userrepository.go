package userrepository

import (
	"fmt"
	"noteservice/model"

	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

type UserStore interface {
	CreateUser(user model.User) error
	GetUser(username string) (model.User, error)
}

func NewUserStore(db *sqlx.DB) UserStore {
	return userRepository{db: db}
}

func (u userRepository) CreateUser(user model.User) error {
	_, err := u.db.Exec("INSERT INTO users VALUES($1,$2)", user.Username, user.Password)
	if err != nil {
		return fmt.Errorf("can't insert into users: %v", err)
	}

	return nil
}

func (u userRepository) GetUser(username string) (model.User, error) {
	row := u.db.QueryRow("SELECT * FROM users WHERE username=$1", username)

	if row.Err() != nil {
		return model.User{}, fmt.Errorf("can't select from users: %v", row.Err())
	}

	user := model.User{}
	var us, pa string
	err := row.Scan(&us, &pa)
	user.Username = us
	user.Password = pa
	if err != nil {
		return model.User{}, fmt.Errorf("can't scan user: %v", row.Err())
	}

	return user, nil
}
