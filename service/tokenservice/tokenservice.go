package tokenservice

import (
	"errors"
	"fmt"
	"noteservice/model"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	AccessToken = iota
	RefreshToken

	ACCESS_TOKEN_TTL  = 15
	REFRESH_TOKEN_TTL = 60
)

type TokenManager interface {
	CreateToken(userinfo model.UserInfo, kind int) (string, error)
	ParseToken(inputToken string, kind int) (model.UserClaims, error)
}

type authService struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewTokenManager(accessSecret, refreshSecret []byte) TokenManager {
	return authService{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
	}
}

func (a authService) CreateToken(userinfo model.UserInfo, kind int) (string, error) {
	claims := model.UserClaims{
		Username: userinfo.Username,
	}

	var secret []byte
	switch kind {
	case AccessToken:
		secret = a.accessSecret
		claims.RegisteredClaims = jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(ACCESS_TOKEN_TTL * time.Minute))}
	case RefreshToken:
		secret = a.refreshSecret
		claims.RegisteredClaims = jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(REFRESH_TOKEN_TTL * time.Minute))}
	default:
		return "", fmt.Errorf("unknown secret kind %d", kind)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secret)
}

func (a authService) ParseToken(inputToken string, kind int) (model.UserClaims, error) {
	token, err := jwt.Parse(inputToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}

		var secret []byte
		switch kind {
		case AccessToken:
			secret = a.accessSecret
		case RefreshToken:
			secret = a.refreshSecret
		default:
			return "", fmt.Errorf("unknown secret kind %d", kind)
		}

		return secret, nil
	})

	if err != nil {
		return model.UserClaims{}, fmt.Errorf("can't parse token: %v", err)
	}

	if !token.Valid {
		return model.UserClaims{}, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return model.UserClaims{}, fmt.Errorf("can't get user claims from token")
	}

	return model.UserClaims{
		Username: claims["username"].(string),
	}, nil
}
