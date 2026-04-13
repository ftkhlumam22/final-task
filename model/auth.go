package model

type AuthResult struct {
	User         User
	AccessToken  string
	RefreshToken string
}
