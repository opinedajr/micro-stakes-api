package auth

type RegisterInput struct {
	FirstName string `json:"first_name" binding:"required,max=100"`
	LastName  string `json:"last_name" binding:"required,max=100"`
	Email     string `json:"email" binding:"required,email,max=255"`
	Password  string `json:"password" binding:"required,min=8"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RegisterOutput struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"fullname"`
	Message  string `json:"message"`
}

type AuthOutput struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
}

type LogoutInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutOutput struct {
	Message string `json:"message"`
}

type ErrorOutput struct {
	Error   string              `json:"error"`
	Code    string              `json:"code"`
	Details map[string][]string `json:"details,omitempty"`
}
