package topicdb

import (
	"time"

	"github.com/dudakovict/gotify/internal/core/topic"
	"github.com/google/uuid"
)

// Topic represents information about an individual topic.
type dbTopic struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

func toDBTopic(tpc topic.Topic) dbTopic {
	return dbTopic{
		ID:        tpc.ID,
		Name:      tpc.Name,
		CreatedAt: tpc.CreatedAt.UTC(),
	}
}

func toCoreTopic(dbTpc dbTopic) topic.Topic {
	return topic.Topic{
		ID:        dbTpc.ID,
		Name:      dbTpc.Name,
		CreatedAt: dbTpc.CreatedAt,
	}
}

func toCoreTopicSlice(dbTpcs []dbTopic) []topic.Topic {
	tpcs := make([]topic.Topic, len(dbTpcs))

	for i, dbTpc := range dbTpcs {
		tpcs[i] = toCoreTopic(dbTpc)
	}

	return tpcs
}
