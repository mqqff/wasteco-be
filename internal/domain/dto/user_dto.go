package dto

import "github.com/google/uuid"

type UserProfile struct {
	ID    string
	Email string
}
type UserParam struct {
	Email string
	ID    uuid.UUID
}
