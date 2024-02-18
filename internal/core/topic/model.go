package topic

import (
	"time"

	"github.com/google/uuid"
)

// Topic represents information about an individual topic.
type Topic struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
}

// NewTopic contains information needed to create a new topic.
type NewTopic struct {
	Name string
}
