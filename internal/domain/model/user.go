package model

import (
	"auth-todo-service/internal/domain"
	"github.com/go-playground/validator/v10"
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleUser UserRole = "user"
)

type User struct {
	ID           uuid.UUID `db:"id" json:"id"`
	FullName     string    `db:"full_name" json:"full_name"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"` // не отдаем
	Role         UserRole  `db:"role" json:"role"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

func NewUser(fullName, email, passwordHash string) (*User, error) {
	user := &User{
		ID:           uuid.New(),
		FullName:     fullName,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         RoleUser,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *User) Validate() error {
	if len(u.FullName) < 2 || len(u.FullName) > 100 {
		return domain.ErrNameInvalid
	}
	if !isValidEmail(u.Email) {
		return domain.ErrInvalidEmail
	}
	if u.PasswordHash == "" {
		return domain.ErrPasswordRequired
	}
	return nil
}

func isValidEmail(email string) bool {
	validate := validator.New()
	err := validate.Var(email, "required,email")
	return err == nil
}
