package dto

import "github.com/google/uuid"

type UserParam struct {
	Email string
	ID    uuid.UUID
}
