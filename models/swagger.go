package models

// UserRequest represents the user input for registration and login
type UserRequest struct {
	Username string `json:"username" example:"john_doe" binding:"required"`
	Password string `json:"password" example:"password123" binding:"required"`
}

// UserResponse represents the user response
type UserResponse struct {
	ID       uint   `json:"id" example:"1"`
	Username string `json:"username" example:"john_doe"`
}

// ProjectRequest represents the project input
type ProjectRequest struct {
	Name string `json:"name" example:"My Project" binding:"required"`
}

// ProjectResponse represents the project response
type ProjectResponse struct {
	ID    uint           `json:"id" example:"1"`
	Name  string         `json:"name" example:"My Project"`
	Tasks []TaskResponse `json:"tasks"`
}

// TaskRequest represents the task input
type TaskRequest struct {
	Title     string `json:"title" example:"Complete feature" binding:"required"`
	Status    string `json:"status" example:"todo" binding:"required"`
	ProjectID uint   `json:"project_id" example:"1" binding:"required"`
}

// TaskResponse represents the task response
type TaskResponse struct {
	ID        uint   `json:"id" example:"1"`
	Title     string `json:"title" example:"Complete feature"`
	Status    string `json:"status" example:"todo"`
	ProjectID uint   `json:"project_id" example:"1"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message" example:"Success"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Something went wrong"`
}