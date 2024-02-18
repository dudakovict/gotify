package notification

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents information about an individual notification.
type Notification struct {
	ID        uuid.UUID
	TopicID   uuid.UUID
	Message   string
	CreatedAt time.Time
}

// NewNotification contains information needed to create a new notification.
type NewNotification struct {
	TopicID uuid.UUID
	Message string
}

// NewNotificationTx contains information needed to create a new notification
// in a transaction.
type NewNotificationTx struct {
	NewNotification
	AfterCreate func(ntf Notification) error
}
