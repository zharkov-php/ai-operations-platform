package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

type memoryStore struct {
	user     User
	tokens   map[[32]byte]bool
	failures int
}

func (m *memoryStore) FindUserByEmail(context.Context, string) (User, error) {
	if m.user.ID == "" {
		return User{}, errors.New("missing")
	}
	return m.user, nil
}
func (m *memoryStore) RecordLoginFailure(context.Context, string, time.Time) error {
	m.failures++
	return nil
}
func (m *memoryStore) RecordLoginSuccess(context.Context, string) error { return nil }
func (m *memoryStore) CreateRefreshToken(_ context.Context, _ User, h [32]byte, _ time.Time) error {
	m.tokens[h] = true
	return nil
}
func (m *memoryStore) RotateRefreshToken(_ context.Context, old, new [32]byte, _ time.Time) (User, error) {
	if !m.tokens[old] {
		return User{}, ErrInvalidToken
	}
	delete(m.tokens, old)
	m.tokens[new] = true
	return m.user, nil
}
func (m *memoryStore) RevokeRefreshToken(_ context.Context, h [32]byte) error {
	if !m.tokens[h] {
		return ErrInvalidToken
	}
	delete(m.tokens, h)
	return nil
}
func (m *memoryStore) Audit(context.Context, User, string) error { return nil }

func testService(t *testing.T) (*Service, *memoryStore) {
	t.Helper()
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatal(err)
	}
	store := &memoryStore{user: User{ID: "user-1", OrganizationID: "org-1", Email: "owner@example.test", Role: "owner", Status: "active", PasswordHash: hash}, tokens: map[[32]byte]bool{}}
	return NewService(store, []byte("test-only-signing-key-with-32-bytes-minimum")), store
}

func TestPasswordHashDoesNotContainPassword(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "correct-horse-battery-staple" {
		t.Fatal("password stored in plaintext")
	}
}
func TestPasswordMinimumLength(t *testing.T) {
	if _, err := HashPassword("short"); err == nil {
		t.Fatal("expected validation error")
	}
}
func TestLoginAndRefreshRotation(t *testing.T) {
	service, _ := testService(t)
	session, err := service.Login(context.Background(), "owner@example.test", "correct-horse-battery-staple")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := service.ParseAccess(session.AccessToken)
	if err != nil || claims.OrganizationID != "org-1" || claims.Role != "owner" {
		t.Fatalf("claims=%+v err=%v", claims, err)
	}
	next, err := service.Refresh(context.Background(), session.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Refresh(context.Background(), session.RefreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("rotated token was reused")
	}
	if next.RefreshToken == session.RefreshToken {
		t.Fatal("refresh token did not rotate")
	}
}
func TestLoginFailureIsGenericAndRecorded(t *testing.T) {
	service, store := testService(t)
	_, err := service.Login(context.Background(), "owner@example.test", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) || store.failures != 1 {
		t.Fatalf("err=%v failures=%d", err, store.failures)
	}
}
func TestTamperedAccessTokenRejected(t *testing.T) {
	service, _ := testService(t)
	session, _ := service.Login(context.Background(), "owner@example.test", "correct-horse-battery-staple")
	if _, err := service.ParseAccess(session.AccessToken + "x"); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("tampered token accepted")
	}
}
