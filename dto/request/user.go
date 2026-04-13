package request

type RegisterUser struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshToken struct {
	RefreshToken string `json:"refresh_token"`
}
