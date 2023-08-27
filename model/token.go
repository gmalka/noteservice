package model

type Refresh struct {
	RefreshToken string `json:"refresh_token"`
}

type Tokens struct {
	AccessToken  string	`json:"access_token"`
	RefreshToken string	`json:"refresh_token"`
}