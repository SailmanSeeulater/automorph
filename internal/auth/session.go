package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/SailmanSeeulater/automorph/internal/store"
)

const (
	CookieName = "automorpth_session"
	SessionTTL = 30 * 24 * time.Hour
)

var ErrInvalidCredentials = errors.New("invalid email or password")

// dummyHash is checked against when the mail doesn't exist, so a wrong email
// takes as long as a wrong password and response times don't reveal which emails
// exist.
var dummyHash, _ = HashPassword("not-a-real-password")

type Service struct {
	q *store.Queries
}

func NewService(q *store.Queries) *Service {
	return &Service{q: q}
}

// Session is what Login hands back to the HTTP layer to put in a cookie.
type Session struct {
	Token     string
	User      store.User
	ExpiresAt time.Time
}

func (s *Service) Login(ctx context.Context, email, password string) (Session, error) {
	user, err := s.q.GetUserByEmail(ctx, strings.TrimSpace(email))
	if errors.Is(err, sql.ErrNoRows) {
		VerifyPassword(password, dummyHash)
		return Session{}, ErrInvalidCredentials
	}
	if err != nil {
		return Session{}, err
	}
	ok, err := VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return Session{}, err
	}
	if !ok {
		return Session{}, ErrInvalidCredentials
	}

	// Housekeeping: logins are rare, so this is a cheap place to clear out old sessions.
	_ = s.q.DeleteExpiredSessions(ctx)

	token := rand.Text()
	expires := time.Now().Add(SessionTTL)
	err = s.q.CreateSession(ctx, store.CreateSessionParams{
		ID:        hashToken(token),
		UserID:    user.ID,
		ExpiresAt: expires.Unix(),
	})
	if err != nil {
		return Session{}, err
	}
	return Session{Token: token, User: user, ExpiresAt: expires}, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	return s.q.DeleteSession(ctx, hashToken(token))
}

// UserForToken returns the user owning an unexpired session, or sql.ErrNoRows.
func (s *Service) UserForToken(ctx context.Context, token string) (store.User, error) {
	row, err := s.q.GetSessionUser(ctx, hashToken(token))
	return row.User, err
}

// Only the hash is stored, so the database alone can't be used to hiject a session.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
