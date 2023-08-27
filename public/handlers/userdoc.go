package handlers

import "noteservice/model"

// swagger:route POST /signin user SigninRequest
// Получить токены аутентификации.
// responses:
//   200: SuccessSignIn
//	 400: BadInputDataResponse
//	 403: AuthErrorResponse
//   500: ServerErrorResponse

// swagger:parameters SigninRequest
type SigninRequest struct {
	// in:body
	Data model.User `json:"auth_params"`
}

// swagger:response SuccessSignIn
type SuccessSignIn struct {
	// in:body
	Tokens model.Tokens `json:"tokens"`
}

// swagger:route POST /signup user SignupRequest
// Зарегестрировать нового пользователя.
// responses:
//   200: SuccessSignUp
//	 400: BadInputDataResponse
//   500: ServerErrorResponse

// swagger:parameters SignupRequest
type SignupRequest struct {
	// in:body
	Data model.User `json:"register_params"`
}

// swagger:response SuccessSignUp
type SuccessSignUp struct {
	// in:body
	Data model.Message	`json:"data"`
}

// swagger:route POST /refresh user RefreshRequest
// Обновить токены.
// responses:
//   200: SuccessRefresh
//	 400: BadInputDataResponse
//	 401: UnauthorizedResponse
//   500: ServerErrorResponse

// swagger:parameters RefreshRequest
type RefreshRequest struct {
	// in:body
	Data model.Refresh `json:"refresh_token"`
}

// swagger:response SuccessRefresh
type SuccessRefresh struct {
	// in:body
	Tokens model.Tokens `json:"tokens"`
}

// swagger:response UnauthorizedResponse
type UnauthorizedResponse struct {
	// in:body
	Message string `json:"message"`
}

// swagger:response BadInputDataResponse
type BadInputDataResponse struct {
	// in:body
	Message string `json:"message"`
}

// swagger:response AuthErrorResponse
type AuthErrorResponse struct {
	// in:body
	Message string `json:"message"`
}

// swagger:response ServerErrorResponse
type ServerErrorResponse struct {
	// in:body
	Message string `json:"message"`
}