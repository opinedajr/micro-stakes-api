package auth

type RegisterInput struct {
	FirstName string `json:"first_name" binding:"required,max=100"`
	LastName  string `json:"last_name" binding:"required,max=100"`
	Email     string `json:"email" binding:"required,email,max=255"`
	Password  string `json:"password" binding:"required,min=8"`
}

type RegisterOutput struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"fullname"`
	Message  string `json:"message"`
}

type ErrorOutput struct {
	Error   string              `json:"error"`
	Code    string              `json:"code"`
	Details map[string][]string `json:"details,omitempty"`
}
