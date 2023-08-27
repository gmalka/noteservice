package postgres

import (
	"fmt"
	"noteservice/database"

	"github.com/jmoiron/sqlx"
)

func NewPostgresConnect(config database.DbConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", config.Host, config.Port, config.User, config.Password, config.Dbname, config.Sslmode))
	if err != nil {
		return nil, fmt.Errorf("can't connect to bd: %v", err)
	}

	return db, nil
}
