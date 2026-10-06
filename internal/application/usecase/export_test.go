package usecase

import (
	"time"

	"github.com/histopathai/main-service/internal/port"
	"golang.org/x/crypto/bcrypt"
)

// NewBlindTestInviteUseCaseForTest hashes PINs at bcrypt's lowest cost and uses the given clock.
func NewBlindTestInviteUseCaseForTest(invites port.BlindTestInviteStore, tests port.BlindTestStore, now func() time.Time) *BlindTestInviteUseCase {
	uc := newBlindTestInviteUseCase(invites, tests, bcrypt.MinCost)
	uc.now = now
	return uc
}
