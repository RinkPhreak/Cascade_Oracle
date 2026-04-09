package dto

// AuthLoginRequest model for operator authentication payload
type AuthLoginRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// AuthTokenResponse model containing standard JWT
type AuthTokenResponse struct {
	AccessToken string `json:"access_token"`
}
