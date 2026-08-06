package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidToken = errors.New("invalid refresh token")

type User struct {
	ID, OrganizationID, Email, DisplayName, Role, Status, PasswordHash string
	FailedLogins                                                       int
	LockedUntil                                                        *time.Time
}
type Session struct {
	User                              User
	AccessToken, RefreshToken         string
	AccessExpiresAt, RefreshExpiresAt time.Time
}
type Claims struct {
	UserID         string    `json:"user_id"`
	OrganizationID string    `json:"organization_id"`
	Role           string    `json:"role"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type Store interface {
	FindUserByEmail(context.Context, string) (User, error)
	RecordLoginFailure(context.Context, string, time.Time) error
	RecordLoginSuccess(context.Context, string) error
	CreateRefreshToken(context.Context, User, [32]byte, time.Time) error
	RotateRefreshToken(context.Context, [32]byte, [32]byte, time.Time) (User, error)
	RevokeRefreshToken(context.Context, [32]byte) error
	Audit(context.Context, User, string) error
}

type Service struct {
	store                 Store
	now                   func() time.Time
	accessTTL, refreshTTL time.Duration
	signingKey            []byte
}

func NewService(store Store, signingKey []byte) *Service {
	return &Service{store: store, now: time.Now, accessTTL: 15 * time.Minute, refreshTTL: 30 * 24 * time.Hour, signingKey: append([]byte(nil), signingKey...)}
}

func HashPassword(password string) (string, error) {
	if len(password) < 12 {
		return "", errors.New("password must contain at least 12 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (s *Service) Login(ctx context.Context, email, password string) (Session, error) {
	user, err := s.store.FindUserByEmail(ctx, email)
	now := s.now().UTC()
	if err != nil || user.Status != "active" || (user.LockedUntil != nil && user.LockedUntil.After(now)) || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		if err == nil {
			_ = s.store.RecordLoginFailure(ctx, user.ID, now)
		}
		return Session{}, ErrInvalidCredentials
	}
	if err := s.store.RecordLoginSuccess(ctx, user.ID); err != nil {
		return Session{}, err
	}
	return s.issue(ctx, user, now, true)
}

func (s *Service) Refresh(ctx context.Context, refresh string) (Session, error) {
	oldHash := sha256.Sum256([]byte(refresh))
	newRefresh, newHash, err := token()
	if err != nil {
		return Session{}, err
	}
	now := s.now().UTC()
	expires := now.Add(s.refreshTTL)
	user, err := s.store.RotateRefreshToken(ctx, oldHash, newHash, expires)
	if err != nil {
		return Session{}, ErrInvalidToken
	}
	access, err := s.encodeAccess(user, now.Add(s.accessTTL))
	if err != nil {
		return Session{}, err
	}
	return Session{User: user, AccessToken: access, RefreshToken: newRefresh, AccessExpiresAt: now.Add(s.accessTTL), RefreshExpiresAt: expires}, nil
}

func (s *Service) Logout(ctx context.Context, refresh string) error {
	hash := sha256.Sum256([]byte(refresh))
	return s.store.RevokeRefreshToken(ctx, hash)
}

func (s *Service) issue(ctx context.Context, user User, now time.Time, audit bool) (Session, error) {
	refresh, hash, err := token()
	if err != nil {
		return Session{}, err
	}
	refreshExpiry := now.Add(s.refreshTTL)
	if err := s.store.CreateRefreshToken(ctx, user, hash, refreshExpiry); err != nil {
		return Session{}, err
	}
	accessExpiry := now.Add(s.accessTTL)
	access, err := s.encodeAccess(user, accessExpiry)
	if err != nil {
		return Session{}, err
	}
	if audit {
		_ = s.store.Audit(ctx, user, "auth.login")
	}
	return Session{User: user, AccessToken: access, RefreshToken: refresh, AccessExpiresAt: accessExpiry, RefreshExpiresAt: refreshExpiry}, nil
}

func token() (string, [32]byte, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", [32]byte{}, err
	}
	value := base64.RawURLEncoding.EncodeToString(raw[:])
	return value, sha256.Sum256([]byte(value)), nil
}

func (s *Service) encodeAccess(user User, expires time.Time) (string, error) {
	payload, err := json.Marshal(Claims{UserID: user.ID, OrganizationID: user.OrganizationID, Role: user.Role, ExpiresAt: expires})
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, s.signingKey)
	_, _ = mac.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func (s *Service) ParseAccess(value string) (Claims, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return Claims{}, ErrInvalidToken
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	mac := hmac.New(sha256.New, s.signingKey)
	_, _ = mac.Write([]byte(parts[0]))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return Claims{}, ErrInvalidToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	var claims Claims
	if json.Unmarshal(payload, &claims) != nil || !claims.ExpiresAt.After(s.now().UTC()) {
		return Claims{}, ErrInvalidToken
	}
	return claims, nil
}
