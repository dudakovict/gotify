package verificationdb

import (
	"time"

	"github.com/dudakovict/gotify/internal/core/verification"
	"github.com/google/uuid"
)

type dbVerification struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Email     string    `db:"email"`
	Code      string    `db:"code"`
	Used      bool      `db:"used"`
	CreatedAt time.Time `db:"created_at"`
	ExpiredAt time.Time `db:"expired_at"`
}

func toDBVerification(vrf verification.Verificiation) dbVerification {
	return dbVerification{
		ID:        vrf.ID,
		UserID:    vrf.UserID,
		Email:     vrf.Email,
		Code:      vrf.Code,
		Used:      vrf.Used,
		CreatedAt: vrf.CreatedAt,
		ExpiredAt: vrf.ExpiredAt,
	}
}

func toCoreVerification(dbVrf dbVerification) verification.Verificiation {
	return verification.Verificiation{
		ID:        dbVrf.ID,
		UserID:    dbVrf.UserID,
		Email:     dbVrf.Email,
		Code:      dbVrf.Code,
		Used:      dbVrf.Used,
		CreatedAt: dbVrf.CreatedAt,
		ExpiredAt: dbVrf.ExpiredAt,
	}
}
