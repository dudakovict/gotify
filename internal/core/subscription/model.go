package subscription

import (
	"time"

	"github.com/google/uuid"
)

// Subscription represents information about an individual subscription.
type Subscription struct {
	ID        uuid.UUID
	TopicID   uuid.UUID
	UserID    uuid.UUID
	CreatedAt time.Time
}

// NewSubscription contains information needed to create a new subscription.
type NewSubscription struct {
	TopicID uuid.UUID
	UserID  uuid.UUID
}
