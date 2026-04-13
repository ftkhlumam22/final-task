package response

import "time"

type LoginUser struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RegisterUser struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
