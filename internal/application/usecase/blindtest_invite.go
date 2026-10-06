package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	stderrors "errors"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	"golang.org/x/crypto/bcrypt"
)

const (
	// BlindTestGuestRole is the role recorded for people who joined through an invitation.
	BlindTestGuestRole = "guest"

	blindTestGuestPrefix     = "guest_"
	maxInviteParticipants    = 500
	maxGuestPINAttempts      = 5
	guestPINLock             = 15 * time.Minute
	maxGuestSessions         = 5
	maxGuestNameLen          = 80
	maxGuestInstitutionLen   = 120
	maxGuestExperienceYears  = 70
	inviteTokenBytes         = 24
	guestSessionSecretBytes  = 32
	guestIDInvitePrefixChars = 12
)

var pinPattern = regexp.MustCompile(`^[0-9]{4}$`)

// BlindTestGuestUserID is the user id a guest's answers are kept under.
func BlindTestGuestUserID(g port.BlindTestGuest) string { return blindTestGuestPrefix + g.ID }

// BlindTestInviteUseCase runs the shared invitation links of the blind test.
type BlindTestInviteUseCase struct {
	invites port.BlindTestInviteStore
	tests   port.BlindTestStore
	now     func() time.Time
	random  func([]byte) (int, error)
	pinCost int
	// dummyPIN is compared against for unknown names, so that a wrong name
	// takes as long as a wrong PIN.
	dummyPIN []byte
}

func NewBlindTestInviteUseCase(invites port.BlindTestInviteStore, tests port.BlindTestStore) *BlindTestInviteUseCase {
	return newBlindTestInviteUseCase(invites, tests, bcrypt.DefaultCost)
}

func newBlindTestInviteUseCase(invites port.BlindTestInviteStore, tests port.BlindTestStore, cost int) *BlindTestInviteUseCase {
	dummy, _ := bcrypt.GenerateFromPassword([]byte("0000"), cost)
	return &BlindTestInviteUseCase{invites: invites, tests: tests, now: time.Now, random: rand.Read, pinCost: cost,
		dummyPIN: dummy}
}

// ── Admin ────────────────────────────────────────────────────────────────────

func (uc *BlindTestInviteUseCase) Create(ctx context.Context, setID, adminID string, max int, expiresAt *time.Time) (*port.BlindTestInvite, error) {
	if max < 1 || max > maxInviteParticipants {
		return nil, coded(errors.ErrorTypeValidation, "max_participants must be between 1 and 500", "max_invalid", nil)
	}
	now := uc.now()
	if expiresAt != nil && !expiresAt.After(now) {
		return nil, coded(errors.ErrorTypeValidation, "expires_at must be in the future", "expiry_invalid", nil)
	}
	set, err := uc.tests.GetSet(ctx, setID)
	if err != nil {
		return nil, err
	}
	if set == nil || !set.Active {
		return nil, errors.NewNotFoundError("blind test not found")
	}
	token, err := uc.secret(inviteTokenBytes)
	if err != nil {
		return nil, err
	}
	inv := port.BlindTestInvite{ID: hashHex(token), Token: token, SetID: setID, MaxParticipants: max,
		ExpiresAt: expiresAt, Active: true, CreatedBy: adminID, CreatedAt: now}
	if err := uc.invites.CreateInvite(ctx, inv); err != nil {
		return nil, err
	}
	return &inv, nil
}

func (uc *BlindTestInviteUseCase) List(ctx context.Context, setID string) ([]port.BlindTestInvite, error) {
	set, err := uc.tests.GetSet(ctx, setID)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return nil, errors.NewNotFoundError("blind test not found")
	}
	list, err := uc.invites.ListInvites(ctx, setID)
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.After(list[j].CreatedAt) })
	return list, nil
}

// Update opens or closes a link and changes its cap; the cap cannot go below
// the people who already joined.
func (uc *BlindTestInviteUseCase) Update(ctx context.Context, setID, inviteID string, active *bool, max *int) (*port.BlindTestInvite, error) {
	inv, err := uc.invites.GetInvite(ctx, inviteID)
	if err != nil {
		return nil, err
	}
	if inv == nil || inv.SetID != setID {
		return nil, errors.NewNotFoundError("invitation not found")
	}
	return uc.invites.UpdateInvite(ctx, inviteID, func(inv *port.BlindTestInvite) error {
		if active != nil {
			inv.Active = *active
		}
		if max != nil {
			if *max < 1 || *max > maxInviteParticipants {
				return coded(errors.ErrorTypeValidation, "max_participants must be between 1 and 500", "max_invalid", nil)
			}
			if *max < inv.Participants {
				return coded(errors.ErrorTypeValidation, "max_participants is below the people who already joined",
					"max_below_participants", map[string]interface{}{"participants": inv.Participants})
			}
			inv.MaxParticipants = *max
		}
		return nil
	})
}

// ── Guests ───────────────────────────────────────────────────────────────────

func (uc *BlindTestInviteUseCase) Info(ctx context.Context, token string) (*port.BlindTestInviteInfo, error) {
	inv, set, err := uc.open(ctx, token)
	if err != nil {
		return nil, err
	}
	info := &port.BlindTestInviteInfo{Participants: inv.Participants, Max: inv.MaxParticipants,
		Closed: !inv.Active || set == nil || !set.Active, Expired: inv.ExpiresAt != nil && uc.now().After(*inv.ExpiresAt)}
	if set != nil {
		info.SetName, info.Description, info.Images = set.Name, set.Description, len(set.ImageIDs)
	}
	info.Joinable = !info.Closed && !info.Expired && inv.Participants < inv.MaxParticipants
	return info, nil
}

// Join admits a new person under a name not yet taken in this invitation.
func (uc *BlindTestInviteUseCase) Join(ctx context.Context, token string, j port.BlindTestJoin) (*port.BlindTestGuestSession, error) {
	name, key, err := guestName(j.Name)
	if err != nil {
		return nil, err
	}
	if !pinPattern.MatchString(j.PIN) {
		return nil, coded(errors.ErrorTypeValidation, "PIN must be 4 digits", "pin_invalid", nil)
	}
	institution := strings.Join(strings.Fields(j.Institution), " ")
	if utf8.RuneCountInString(institution) > maxGuestInstitutionLen {
		return nil, coded(errors.ErrorTypeValidation, "institution is too long", "institution_invalid", nil)
	}
	if j.ExperienceYears != nil && (*j.ExperienceYears < 0 || *j.ExperienceYears > maxGuestExperienceYears) {
		return nil, coded(errors.ErrorTypeValidation, "experience_years must be between 0 and 70", "experience_invalid", nil)
	}
	if !j.Consent {
		return nil, coded(errors.ErrorTypeValidation, "consent is required", "consent_required", nil)
	}

	inv, set, err := uc.open(ctx, token)
	if err != nil {
		return nil, err
	}
	now := uc.now()
	switch {
	case !inv.Active || set == nil || !set.Active:
		return nil, coded(errors.ErrorTypeForbidden, "this invitation is closed", "invite_closed", nil)
	case inv.ExpiresAt != nil && now.After(*inv.ExpiresAt):
		return nil, coded(errors.ErrorTypeForbidden, "this invitation has expired", "invite_expired", nil)
	}

	pinHash, err := bcrypt.GenerateFromPassword([]byte(j.PIN), uc.pinCost)
	if err != nil {
		return nil, errors.NewInternalError("could not hash the PIN", err)
	}
	session, sessionHash, err := uc.newSession(guestID(inv.ID, key))
	if err != nil {
		return nil, err
	}
	g := port.BlindTestGuest{ID: guestID(inv.ID, key), InviteID: inv.ID, SetID: inv.SetID, Name: name, NameKey: key,
		Institution: institution, ExperienceYears: j.ExperienceYears, PinHash: string(pinHash),
		SessionHashes: []string{sessionHash}, ConsentAt: now, CreatedAt: now}
	switch err := uc.invites.AddGuest(ctx, g, inv.MaxParticipants); {
	case stderrors.Is(err, port.ErrBlindTestNameTaken):
		return nil, coded(errors.ErrorTypeConflict, "this name already joined; continue with its PIN", "name_taken", nil)
	case stderrors.Is(err, port.ErrBlindTestInviteFull):
		return nil, coded(errors.ErrorTypeConflict, "this invitation is full", "invite_full", nil)
	case err != nil:
		return nil, err
	}
	return &port.BlindTestGuestSession{Token: session, Guest: g}, nil
}

// Resume lets a person back in with their name and PIN, on any device. Five
// wrong PINs lock the name for 15 minutes. Allowed after the link expired
// (expiry only stops new people), not after it was closed.
func (uc *BlindTestInviteUseCase) Resume(ctx context.Context, token, name, pin string) (*port.BlindTestGuestSession, error) {
	inv, set, err := uc.open(ctx, token)
	if err != nil {
		return nil, err
	}
	if !inv.Active || set == nil || !set.Active {
		return nil, coded(errors.ErrorTypeForbidden, "this invitation is closed", "invite_closed", nil)
	}
	wrong := coded(errors.ErrorTypeUnauthorized, "name or PIN is wrong", "wrong_credentials", nil)
	_, key, err := guestName(name)
	if err != nil {
		_ = bcrypt.CompareHashAndPassword(uc.dummyPIN, []byte(pin))
		return nil, wrong
	}
	g, err := uc.invites.GetGuest(ctx, guestID(inv.ID, key))
	if err != nil {
		return nil, err
	}
	if g == nil || g.InviteID != inv.ID {
		_ = bcrypt.CompareHashAndPassword(uc.dummyPIN, []byte(pin))
		return nil, wrong
	}
	now := uc.now()
	if g.LockedUntil != nil && now.Before(*g.LockedUntil) {
		return nil, coded(errors.ErrorTypeForbidden, "too many wrong PINs; try again later", "locked",
			map[string]interface{}{"locked_until": g.LockedUntil.UTC().Format(time.RFC3339)})
	}

	if bcrypt.CompareHashAndPassword([]byte(g.PinHash), []byte(pin)) != nil {
		updated, err := uc.invites.UpdateGuest(ctx, g.ID, func(g *port.BlindTestGuest) error {
			g.FailedAttempts++
			if g.FailedAttempts >= maxGuestPINAttempts {
				until := now.Add(guestPINLock)
				g.LockedUntil, g.FailedAttempts = &until, 0
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		if updated.LockedUntil != nil && now.Before(*updated.LockedUntil) {
			return nil, coded(errors.ErrorTypeForbidden, "too many wrong PINs; try again later", "locked",
				map[string]interface{}{"locked_until": updated.LockedUntil.UTC().Format(time.RFC3339)})
		}
		wrong.Details["attempts_left"] = maxGuestPINAttempts - updated.FailedAttempts
		return nil, wrong
	}

	session, sessionHash, err := uc.newSession(g.ID)
	if err != nil {
		return nil, err
	}
	updated, err := uc.invites.UpdateGuest(ctx, g.ID, func(g *port.BlindTestGuest) error {
		g.FailedAttempts, g.LockedUntil = 0, nil
		g.SessionHashes = append(g.SessionHashes, sessionHash)
		if n := len(g.SessionHashes); n > maxGuestSessions {
			g.SessionHashes = g.SessionHashes[n-maxGuestSessions:]
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &port.BlindTestGuestSession{Token: session, Guest: *updated}, nil
}

// Authenticate checks a guest's session token against the invitation.
func (uc *BlindTestInviteUseCase) Authenticate(ctx context.Context, token, session string) (*port.BlindTestGuest, error) {
	invalid := coded(errors.ErrorTypeUnauthorized, "session is not valid", "session_invalid", nil)
	inv, set, err := uc.open(ctx, token)
	if err != nil {
		return nil, err
	}
	if !inv.Active || set == nil || !set.Active {
		return nil, coded(errors.ErrorTypeForbidden, "this invitation is closed", "invite_closed", nil)
	}
	i := strings.LastIndexByte(session, '.')
	if i <= 0 || i == len(session)-1 {
		return nil, invalid
	}
	g, err := uc.invites.GetGuest(ctx, session[:i])
	if err != nil {
		return nil, err
	}
	if g == nil || g.InviteID != inv.ID {
		return nil, invalid
	}
	want := []byte(hashHex(session[i+1:]))
	for _, h := range g.SessionHashes {
		if subtle.ConstantTimeCompare([]byte(h), want) == 1 {
			return g, nil
		}
	}
	return nil, invalid
}

// open finds the invitation of a link token and its set (nil if the set is gone).
func (uc *BlindTestInviteUseCase) open(ctx context.Context, token string) (*port.BlindTestInvite, *port.BlindTestSet, error) {
	if token == "" {
		return nil, nil, errors.NewNotFoundError("invitation not found")
	}
	inv, err := uc.invites.GetInvite(ctx, hashHex(token))
	if err != nil {
		return nil, nil, err
	}
	if inv == nil {
		return nil, nil, errors.NewNotFoundError("invitation not found")
	}
	set, err := uc.tests.GetSet(ctx, inv.SetID)
	if err != nil {
		return nil, nil, err
	}
	return inv, set, nil
}

func (uc *BlindTestInviteUseCase) secret(n int) (string, error) {
	b := make([]byte, n)
	if _, err := uc.random(b); err != nil {
		return "", errors.NewInternalError("could not make a random token", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// newSession returns "guestID.secret" for the guest and the hash to keep.
func (uc *BlindTestInviteUseCase) newSession(guestID string) (token, hash string, err error) {
	secret, err := uc.secret(guestSessionSecretBytes)
	if err != nil {
		return "", "", err
	}
	return guestID + "." + secret, hashHex(secret), nil
}

func hashHex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func guestID(inviteID, nameKey string) string {
	return inviteID[:min(guestIDInvitePrefixChars, len(inviteID))] + "_" + nameKey
}

// guestName tidies a typed name and makes its key: Turkish letters folded to
// ASCII, case ignored, anything else a single dash — "Ayşe  YILMAZ" and
// "ayse yilmaz" are the same person.
func guestName(raw string) (name, key string, err error) {
	name = strings.Join(strings.Fields(raw), " ")
	if n := utf8.RuneCountInString(name); n < 2 || n > maxGuestNameLen {
		return "", "", coded(errors.ErrorTypeValidation, "name must be 2 to 80 characters", "name_invalid", nil)
	}
	fold := map[rune]rune{'ı': 'i', 'İ': 'i', 'I': 'i', 'ş': 's', 'Ş': 's', 'ğ': 'g', 'Ğ': 'g', 'ü': 'u', 'Ü': 'u',
		'ö': 'o', 'Ö': 'o', 'ç': 'c', 'Ç': 'c', 'â': 'a', 'Â': 'a', 'î': 'i', 'Î': 'i', 'û': 'u', 'Û': 'u'}
	var b strings.Builder
	dash := false
	for _, r := range name {
		if f, ok := fold[r]; ok {
			r = f
		}
		r = unicode.ToLower(r)
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
		} else if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	key = strings.TrimSuffix(b.String(), "-")
	if key == "" {
		return "", "", coded(errors.ErrorTypeValidation, "name must contain letters", "name_invalid", nil)
	}
	return name, key, nil
}

// coded is an application error carrying a machine-readable code for the client.
func coded(t errors.ErrorType, message, code string, details map[string]interface{}) *errors.Err {
	if details == nil {
		details = map[string]interface{}{}
	}
	details["code"] = code
	return &errors.Err{Type: t, Message: message, Details: details}
}
