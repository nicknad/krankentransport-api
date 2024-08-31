package main

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

const AdminRole = "Admin"

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginResponse struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

type User struct {
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Name         string `json:"name"`
	Role         string `json:"role"`
}

func (u *User) ValidPassword(pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pw)) == nil
}

func NewUser(email, name, password, role string) (*User, error) {
	encpw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		Email:        email,
		Name:         name,
		PasswordHash: string(encpw),
		Role:         role,
	}, nil
}

type Krankenfahrt struct {
	Id          int64      `json:"id"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"createdAt"`
	AcceptedBy  *string    `json:"acceptedBy"`
	AcceptedAt  *time.Time `json:"acceptedAt"`
	Finished    bool       `json:"finished"`
}
