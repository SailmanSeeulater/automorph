package auth

import (
	"context"
	"net/http"

	"github.com/SailmanSeeulater/automorph/internal/store"
)

type ctxKey struct{}

// LoadUser look up th esession cookie and, if it's valid, puts the user in the
// request context. It never rejects a request: handler decide what needs a login.
func (s *Service) LoadUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie(CookieName); err == nil && c.Value != "" {
			if user, err := s.UserForToken(r.Context(), c.Value); err == nil {
				r = r.WithContext(context.WithValue(r.Context(), ctxKey{}, user))
			}
		}
		next.ServeHTTP(w, r)
	})
}

// UserForm returns the logged-in user, if any.
func UserFrom(ctx context.Context) (store.User, bool) {
	user, ok := ctx.Value(ctxKey{}).(store.User)
	return user, ok
}
