package maker

import (
	"testing"
	"time"

	"github.com/dudakovict/gotify/internal/core/user"
	"github.com/dudakovict/gotify/pkg/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPasetoMaker(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	userID, err := uuid.NewRandom()
	require.NoError(t, err)
	roles := []string{user.RoleUser.Name()}
	duration := time.Minute

	issuedAt := time.Now()
	expiresAt := time.Now().Add(duration)

	token, payload, err := maker.CreateToken(userID, roles, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, userID, payload.UserID)
	require.Equal(t, roles, payload.Roles)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiresAt, payload.ExpiresAt, time.Second)
}

func TestExpiredPasetoToken(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	userID, err := uuid.NewRandom()
	require.NoError(t, err)

	token, payload, err := maker.CreateToken(userID, []string{user.RoleUser.Name()}, -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}
