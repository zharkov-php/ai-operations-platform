package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Client   string `json:"client"`
}
type tokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}
type sessionResponse struct {
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func registerAuthRoutes(router chi.Router, service *auth.Service) {
	if service == nil {
		return
	}
	router.Post("/api/v1/auth/login", login(service))
	router.Post("/api/v1/auth/refresh", refresh(service))
	router.Post("/api/v1/auth/logout", logout(service))
	router.With(authenticate(service)).Get("/api/v1/me", me)
}

func login(service *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input loginRequest
		if json.NewDecoder(r.Body).Decode(&input) != nil || input.Email == "" || input.Password == "" {
			writeError(w, r, 400, "invalid_request", "email and password are required")
			return
		}
		session, err := service.Login(r.Context(), input.Email, input.Password)
		if err != nil {
			writeError(w, r, 401, "invalid_credentials", "email or password is invalid")
			return
		}
		if input.Client == "mobile" {
			writeJSON(w, 200, sessionResponse{AccessToken: session.AccessToken, RefreshToken: session.RefreshToken, TokenType: "Bearer", ExpiresAt: session.AccessExpiresAt})
			return
		}
		setCookie(w, "access_token", session.AccessToken, session.AccessExpiresAt)
		setCookie(w, "refresh_token", session.RefreshToken, session.RefreshExpiresAt)
		writeJSON(w, 200, sessionResponse{ExpiresAt: session.AccessExpiresAt})
	}
}

func refresh(service *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input tokenRequest
		_ = json.NewDecoder(r.Body).Decode(&input)
		mobile := input.RefreshToken != ""
		if input.RefreshToken == "" {
			if cookie, err := r.Cookie("refresh_token"); err == nil {
				input.RefreshToken = cookie.Value
			}
		}
		session, err := service.Refresh(r.Context(), input.RefreshToken)
		if err != nil {
			writeError(w, r, 401, "invalid_refresh_token", "refresh token is invalid")
			return
		}
		if mobile {
			writeJSON(w, 200, sessionResponse{AccessToken: session.AccessToken, RefreshToken: session.RefreshToken, TokenType: "Bearer", ExpiresAt: session.AccessExpiresAt})
			return
		}
		setCookie(w, "access_token", session.AccessToken, session.AccessExpiresAt)
		setCookie(w, "refresh_token", session.RefreshToken, session.RefreshExpiresAt)
		writeJSON(w, 200, sessionResponse{ExpiresAt: session.AccessExpiresAt})
	}
}

func logout(service *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input tokenRequest
		_ = json.NewDecoder(r.Body).Decode(&input)
		if input.RefreshToken == "" {
			if cookie, err := r.Cookie("refresh_token"); err == nil {
				input.RefreshToken = cookie.Value
			}
		}
		if input.RefreshToken != "" {
			_ = service.Logout(r.Context(), input.RefreshToken)
		}
		expireCookie(w, "access_token", "/")
		expireCookie(w, "refresh_token", "/api/v1/auth")
		w.WriteHeader(http.StatusNoContent)
	}
}

type claimsKey struct{}

func authenticate(service *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value := ""
			header := r.Header.Get("Authorization")
			if strings.HasPrefix(header, "Bearer ") {
				value = strings.TrimPrefix(header, "Bearer ")
			} else if cookie, err := r.Cookie("access_token"); err == nil {
				value = cookie.Value
			}
			claims, err := service.ParseAccess(value)
			if err != nil {
				writeError(w, r, 401, "unauthorized", "authentication is required")
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsKey{}, claims)))
		})
	}
}

func requireRoles(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(claimsKey{}).(auth.Claims)
			if !ok {
				writeError(w, r, 401, "unauthorized", "authentication is required")
				return
			}
			if _, ok := allowed[claims.Role]; !ok {
				writeError(w, r, 403, "forbidden", "permission is required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
func me(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(claimsKey{}).(auth.Claims)
	writeJSON(w, 200, map[string]string{"id": claims.UserID, "organization_id": claims.OrganizationID, "role": claims.Role})
}
func setCookie(w http.ResponseWriter, name, value string, expires time.Time) {
	path := "/"
	sameSite := http.SameSiteLaxMode
	if name == "refresh_token" {
		path = "/api/v1/auth"
		sameSite = http.SameSiteStrictMode
	}
	http.SetCookie(w, &http.Cookie{Name: name, Value: value, Path: path, Expires: expires, HttpOnly: true, Secure: true, SameSite: sameSite})
}
func expireCookie(w http.ResponseWriter, name, path string) {
	http.SetCookie(w, &http.Cookie{Name: name, Value: "", Path: path, MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode})
}
