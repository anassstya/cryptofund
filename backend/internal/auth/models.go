package auth

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate=:"required,min=8,max=64"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate=:"required,min=8,nax=64"`
}

type ResponseAuth struct {
	Token 	string 	 `json:"token"`
	UserID	string	 `json:"user_id"`
	Email 	string	 `json:"email"`
}
