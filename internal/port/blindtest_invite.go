package port

import (
	"context"
	stderrors "errors"
	"time"
)

// Blind test invitations: one shared link per set, sent to a group of
// pathologists who have no platform account. Opening it, a person gives their
// name, chooses a 4-digit PIN and takes the test. The same link brings them back: the browser
// remembers them, and on another device their name and PIN do. The admin caps
// the number of people a link admits and can close it.

// BlindTestInvite is a shared invitation link to one set. Admins only.
type BlindTestInvite struct {
	// ID is the SHA-256 (hex) of the link token: the token itself is never a key.
	ID string
	// Token is the secret part of the link, kept so an admin can copy the link again.
	Token           string
	SetID           string
	MaxParticipants int
	// ExpiresAt stops new people joining after it; those who joined may go on.
	ExpiresAt *time.Time
	// Active false closes the link entirely: nobody joins or continues.
	Active    bool
	CreatedBy string
	CreatedAt time.Time
	// Participants is how many people joined (filled when read).
	Participants int
}

// BlindTestGuest is a person who joined through an invitation.
type BlindTestGuest struct {
	// ID is unique per invitation and name: a name is taken once per link.
	ID       string
	InviteID string
	SetID    string
	// Name is stored in capitals with Turkish letters as ASCII ("AYSE YILMAZ").
	Name    string
	NameKey string
	PinHash string
	// SessionHashes are SHA-256 (hex) of the session secrets handed out (one per device, newest last).
	SessionHashes  []string
	FailedAttempts int
	LockedUntil    *time.Time
	ConsentAt      time.Time
	CreatedAt      time.Time
}

// BlindTestGuestProfile is what the results show of a guest participant.
type BlindTestGuestProfile struct {
	Name string
}

var (
	// ErrBlindTestInviteFull: the invitation admitted as many people as it may.
	ErrBlindTestInviteFull = stderrors.New("blind test invitation is full")
	// ErrBlindTestNameTaken: someone already joined this invitation under this name.
	ErrBlindTestNameTaken = stderrors.New("name already taken in this invitation")
)

type BlindTestInviteStore interface {
	CreateInvite(ctx context.Context, invite BlindTestInvite) error
	// GetInvite returns nil, nil when there is none.
	GetInvite(ctx context.Context, id string) (*BlindTestInvite, error)
	ListInvites(ctx context.Context, setID string) ([]BlindTestInvite, error)
	UpdateInvite(ctx context.Context, id string, change func(*BlindTestInvite) error) (*BlindTestInvite, error)
	// AddGuest stores the guest if the invitation has room (fewer than max
	// guests) and the guest's ID is free, atomically; otherwise it returns
	// ErrBlindTestInviteFull or ErrBlindTestNameTaken.
	AddGuest(ctx context.Context, guest BlindTestGuest, max int) error
	// GetGuest returns nil, nil when there is none.
	GetGuest(ctx context.Context, id string) (*BlindTestGuest, error)
	UpdateGuest(ctx context.Context, id string, change func(*BlindTestGuest) error) (*BlindTestGuest, error)
	ListGuests(ctx context.Context, setID string) ([]BlindTestGuest, error)
}

// BlindTestInviteInfo is what anyone holding the link may know before joining.
// No description: guests learn nothing of the set beyond its neutral name.
type BlindTestInviteInfo struct {
	SetName      string
	Images       int
	Participants int
	Max          int
	// Joinable: open, not expired, not full. Closed: the admin closed the link.
	Joinable bool
	Closed   bool
	Expired  bool
}

type BlindTestJoin struct {
	Name    string
	PIN     string
	Consent bool
}

// BlindTestGuestSession is handed to a guest on join or resume; Token goes in
// the X-Guest-Session header of their test requests.
type BlindTestGuestSession struct {
	Token string
	Guest BlindTestGuest
}

type BlindTestInviteUseCase interface {
	Create(ctx context.Context, setID, adminID string, max int, expiresAt *time.Time) (*BlindTestInvite, error)
	List(ctx context.Context, setID string) ([]BlindTestInvite, error)
	Update(ctx context.Context, setID, inviteID string, active *bool, max *int) (*BlindTestInvite, error)

	Info(ctx context.Context, token string) (*BlindTestInviteInfo, error)
	Join(ctx context.Context, token string, join BlindTestJoin) (*BlindTestGuestSession, error)
	Resume(ctx context.Context, token, name, pin string) (*BlindTestGuestSession, error)
	// Authenticate returns the guest a session token belongs to, if the link is still open.
	Authenticate(ctx context.Context, token, session string) (*BlindTestGuest, error)
}

// BlindTestGuestLister lets the results name guest participants.
type BlindTestGuestLister interface {
	ListGuests(ctx context.Context, setID string) ([]BlindTestGuest, error)
}
