package domain

import (
	"github.com/google/uuid"
	"time"
)

type RefreshToken struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
	Revoked   bool      `db:"revoked"`
}
