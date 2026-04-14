package model

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	DeletedAt *time.Time
}

type Project struct {
	ID        uuid.UUID
	AccountID uuid.UUID
	Name      string
	Slug      string
	CreatedAt time.Time
	DeletedAt *time.Time
}

type APIKey struct {
	ID         uuid.UUID
	ProjectID  uuid.UUID
	Name       string
	KeyHash    string
	KeyPrefix  string
	CreatedAt  time.Time
	LastUsedAt *time.Time
	RevokedAt  *time.Time
}
