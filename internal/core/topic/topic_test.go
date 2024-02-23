package topic_test

import (
	"database/sql"
	"testing"

	"github.com/dudakovict/gotify/internal/core/topic"
	mocktpc "github.com/dudakovict/gotify/internal/core/topic/mock"
	"github.com/dudakovict/gotify/pkg/util"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	tpc := randomTopic()

	testCases := []struct {
		name          string
		buildStubs    func(storer *mocktpc.MockStorer)
		checkResponse func(err error, createdTpc topic.Topic)
	}{
		{
			name: "OK",
			buildStubs: func(storer *mocktpc.MockStorer) {
				arg := topic.Topic{
					ID:        tpc.ID,
					Name:      tpc.Name,
					CreatedAt: tpc.CreatedAt,
				}

				storer.EXPECT().
					Create(arg).
					Times(1).
					Return(nil)
			},
			checkResponse: func(err error, createdTpc topic.Topic) {
				require.NoError(t, err)
				require.NotNil(t, createdTpc)
				require.Equal(t, createdTpc.Name, tpc.Name)
				require.Equal(t, createdTpc.ID, tpc.ID)
			},
		},
		{
			name: "DuplicateName",
			buildStubs: func(storer *mocktpc.MockStorer) {
				storer.EXPECT().
					Create(gomock.Any()).
					Times(1).
					Return(topic.ErrUniqueName)
			},
			checkResponse: func(err error, createdTpc topic.Topic) {
				require.ErrorIs(t, err, topic.ErrUniqueName)
				require.Empty(t, createdTpc)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storer := mocktpc.NewMockStorer(ctrl)
		core := topic.NewCore(nil, storer, gen)

		tc.buildStubs(storer)

		createdTpc, err := core.Create(topic.NewTopic{
			Name: tpc.Name,
		})

		tc.checkResponse(err, createdTpc)
	}
}

func TestQueryByID(t *testing.T) {
	tpc := randomTopic()

	testCases := []struct {
		name          string
		buildStubs    func(storer *mocktpc.MockStorer)
		checkResponse func(err error, queryTpc topic.Topic)
	}{
		{
			name: "OK",
			buildStubs: func(storer *mocktpc.MockStorer) {
				storer.EXPECT().
					QueryByID(tpc.ID).
					Times(1).
					Return(tpc, nil)
			},
			checkResponse: func(err error, queryTpc topic.Topic) {
				require.NoError(t, err)
				require.Equal(t, tpc.ID, queryTpc.ID)
				require.Equal(t, tpc.Name, queryTpc.Name)
			},
		},
		{
			name: "NotFound",
			buildStubs: func(storer *mocktpc.MockStorer) {
				storer.EXPECT().
					QueryByID(tpc.ID).
					Times(1).
					Return(topic.Topic{}, topic.ErrNotFound)
			},
			checkResponse: func(err error, queryTpc topic.Topic) {
				require.ErrorIs(t, err, topic.ErrNotFound)
				require.Empty(t, queryTpc)
			},
		},
		{
			name: "InternalError",
			buildStubs: func(storer *mocktpc.MockStorer) {
				storer.EXPECT().
					QueryByID(tpc.ID).
					Times(1).
					Return(topic.Topic{}, sql.ErrConnDone)
			},
			checkResponse: func(err error, queryTpc topic.Topic) {
				require.ErrorIs(t, err, sql.ErrConnDone)
				require.Empty(t, queryTpc)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storer := mocktpc.NewMockStorer(ctrl)
		core := topic.NewCore(nil, storer, gen)

		tc.buildStubs(storer)

		queryTpc, err := core.QueryByID(tpc.ID)

		tc.checkResponse(err, queryTpc)
	}
}

func TestQuery(t *testing.T) {
	pageNumber := 1
	rowsPerPage := 10

	tpcs := []topic.Topic{
		randomTopic(),
		randomTopic(),
		randomTopic(),
	}

	testCases := []struct {
		name          string
		buildStubs    func(storer *mocktpc.MockStorer)
		checkResponse func(err error, queryTpcs []topic.Topic)
	}{
		{
			name: "OK",
			buildStubs: func(storer *mocktpc.MockStorer) {
				storer.EXPECT().
					Query(pageNumber, rowsPerPage).
					Times(1).
					Return(tpcs, nil)
			},
			checkResponse: func(err error, queryTpcs []topic.Topic) {
				require.NoError(t, err)
				require.Equal(t, len(tpcs), len(queryTpcs))
				require.Equal(t, tpcs, queryTpcs)
			},
		},
		{
			name: "InternalError",
			buildStubs: func(storer *mocktpc.MockStorer) {
				storer.EXPECT().
					Query(pageNumber, rowsPerPage).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(err error, queryTpcs []topic.Topic) {
				require.ErrorIs(t, err, sql.ErrConnDone)
				require.Nil(t, queryTpcs)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storer := mocktpc.NewMockStorer(ctrl)
		core := topic.NewCore(nil, storer, gen)

		tc.buildStubs(storer)

		queryTpcs, err := core.Query(pageNumber, rowsPerPage)

		tc.checkResponse(err, queryTpcs)
	}
}

func randomTopic() topic.Topic {
	return topic.Topic{
		ID:   gen(),
		Name: util.RandomString(6),
	}
}

func gen() uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-000000000000")
}
