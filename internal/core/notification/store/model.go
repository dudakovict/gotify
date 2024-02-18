package notificationdb

import (
	"time"

	"github.com/dudakovict/gotify/internal/core/notification"
	"github.com/google/uuid"
)

type dbNotification struct {
	ID        uuid.UUID `db:"id"`
	TopicID   uuid.UUID `db:"topic_id"`
	Message   string    `db:"message"`
	CreatedAt time.Time `db:"created_at"`
}

func toDBNotification(ntf notification.Notification) dbNotification {
	return dbNotification{
		ID:        ntf.ID,
		TopicID:   ntf.TopicID,
		Message:   ntf.Message,
		CreatedAt: ntf.CreatedAt,
	}
}

func toCoreNotification(dbNtf dbNotification) notification.Notification {
	return notification.Notification{
		ID:        dbNtf.ID,
		TopicID:   dbNtf.TopicID,
		Message:   dbNtf.Message,
		CreatedAt: dbNtf.CreatedAt,
	}
}

func toCoreNotificationSlice(dbNtfs []dbNotification) []notification.Notification {
	ntfs := make([]notification.Notification, len(dbNtfs))

	for i, dbNtf := range dbNtfs {
		ntfs[i] = toCoreNotification(dbNtf)
	}

	return ntfs
}
