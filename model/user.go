package model

import "github.com/golang-jwt/jwt/v4"

type User struct {
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

type UserClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type UserInfo struct {
	Username string
}
