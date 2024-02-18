package verification

import (
	"time"

	"github.com/google/uuid"
)

type Verificiation struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Email     string
	Code      string
	Used      bool
	CreatedAt time.Time
	ExpiredAt time.Time
}

type NewVerification struct {
	UserID uuid.UUID
	Email  string
	Code   string
}
