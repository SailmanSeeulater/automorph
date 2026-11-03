package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/SailmanSeeulater/automorph/internal/auth"
	"github.com/SailmanSeeulater/automorph/internal/store"
)

// UserDTO is the public shape of a user. Never return store.User directly:
// it contains the password hash.
type UserDTO struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role" enum:"admin,member"`
}

func toUserDTO(u store.User) UserDTO {
	return UserDTO{ID: u.ID, Email: u.Email, DisplayName: u.DisplayName, Role: u.Role}
}

type LoginInput struct {
	Body struct {
		Email    string `json:"email" format:"email" maxLength:"254"`
		Password string `json:"password" minLength:"1" maxLength:"256"`
	}
}

type LoginOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      UserDTO
}

type LogoutInput struct {
	Session string `cookie:"automorph_session"`
}

type LogoutOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
}

type MeOutput struct {
	Body UserDTO
}

func registerAuth(api huma.API, d Deps) {
	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/api/auth/login",
		Summary:     "Log in",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, in *LoginInput) (*LoginOutput, error) {
		sess, err := d.Auth.Login(ctx, in.Body.Email, in.Body.Password)
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, huma.Error401Unauthorized("invalid email or password")
		}
		if err != nil {
			return nil, err
		}
		cookie := d.sessionCookie(sess.Token)
		cookie.Expires = sess.ExpiresAt
		return &LoginOutput{SetCookie: cookie, Body: toUserDTO(sess.User)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "logout",
		Method:      http.MethodPost,
		Path:        "/api/auth/logout",
		Summary:     "Log out",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, in *LogoutInput) (*LogoutOutput, error) {
		if in.Session != "" {
			if err := d.Auth.Logout(ctx, in.Session); err != nil {
				return nil, err
			}
		}
		cookie := d.sessionCookie("")
		cookie.MaxAge = -1 // tells the browser to delete it
		return &LogoutOutput{SetCookie: cookie}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-me",
		Method:      http.MethodGet,
		Path:        "/api/me",
		Summary:     "Current user",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, _ *struct{}) (*MeOutput, error) {
		user, ok := auth.UserFrom(ctx)
		if !ok {
			return nil, huma.Error401Unauthorized("not logged in")
		}
		return &MeOutput{Body: toUserDTO(user)}, nil
	})
}

// requireAdmin is the guard for every write endpoint. Used from step 5 on.
func requireAdmin(ctx context.Context) (store.User, error) {
	user, ok := auth.UserFrom(ctx)
	if !ok {
		return store.User{}, huma.Error401Unauthorized("not logged in")
	}
	if user.Role != "admin" {
		return store.User{}, huma.Error403Forbidden("admins only")
	}
	return user, nil
}

func (d Deps) sessionCookie(value string) http.Cookie {
	return http.Cookie{
		Name:     auth.CookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,            // JavaScript can't read it, so XSS can't steal it
		Secure:   d.SecureCookies, // HTTPS only in production
		SameSite: http.SameSiteLaxMode,
	}
}
