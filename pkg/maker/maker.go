// Package maker provides an interface and related functionality for
// token creation and verification.
package maker

import (
	"time"

	"github.com/google/uuid"
)

// Maker is the interface that defines methods for creating and verifying tokens.
type Maker interface {
	CreateToken(id uuid.UUID, roles []string, duration time.Duration) (string, *Payload, error)
	VerifyToken(token string) (*Payload, error)
}
