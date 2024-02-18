package subscriptiondb

import (
	"time"

	"github.com/dudakovict/gotify/internal/core/subscription"
	"github.com/google/uuid"
)

// Subscription represents information about an individual subscription.
type dbSubscription struct {
	ID        uuid.UUID `db:"id"`
	TopicID   uuid.UUID `db:"topic_id"`
	UserID    uuid.UUID `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

func toDBSubscription(sub subscription.Subscription) dbSubscription {
	return dbSubscription{
		ID:        sub.ID,
		TopicID:   sub.TopicID,
		UserID:    sub.UserID,
		CreatedAt: sub.CreatedAt.UTC(),
	}
}

func toCoreSubscription(dbSub dbSubscription) subscription.Subscription {
	return subscription.Subscription{
		ID:        dbSub.ID,
		TopicID:   dbSub.TopicID,
		UserID:    dbSub.UserID,
		CreatedAt: dbSub.CreatedAt,
	}
}

func toCoreSubscriptionSlice(dbSubs []dbSubscription) []subscription.Subscription {
	subs := make([]subscription.Subscription, len(dbSubs))

	for i, dbSub := range dbSubs {
		subs[i] = toCoreSubscription(dbSub)
	}

	return subs
}
