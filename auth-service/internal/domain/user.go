package domain

import (
	"time"
)

// User represents a user in the domain
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never expose password in JSON
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	List(page, limit int) ([]*User, int, error)
	Update(user *User) error
	Delete(id string) error
}

// UserService defines the interface for user business logic
type UserService interface {
	CreateUser(name, email, password string) (*User, error)
	GetUser(id string) (*User, error)
	ListUsers(page, limit int) ([]*User, int, error)
	UpdateUser(id, name, email string) (*User, error)
	DeleteUser(id string) error
	ValidateUser(user *User) error
}
