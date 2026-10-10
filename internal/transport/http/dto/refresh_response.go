package dto

type TypeToken string

type RefreshResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    TypeToken `json:"token_type"`
}
