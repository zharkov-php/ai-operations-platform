package auth

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

func (s *PostgresStore) FindUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := s.pool.QueryRow(ctx, `SELECT u.id, u.organization_id, u.email, u.display_name, u.status, u.password_hash,
        u.failed_login_count, u.locked_until, m.role FROM users u JOIN memberships m ON m.user_id=u.id AND m.organization_id=u.organization_id
        WHERE lower(u.email)=lower($1)`, email).Scan(&user.ID, &user.OrganizationID, &user.Email, &user.DisplayName, &user.Status, &user.PasswordHash, &user.FailedLogins, &user.LockedUntil, &user.Role)
	return user, err
}

func (s *PostgresStore) RecordLoginFailure(ctx context.Context, userID string, now time.Time) error {
	_, err := s.pool.Exec(ctx, `UPDATE users SET failed_login_count=failed_login_count+1,
        locked_until=CASE WHEN failed_login_count+1 >= 5 THEN $2 + interval '15 minutes' ELSE locked_until END,
        updated_at=$2 WHERE id=$1`, userID, now)
	return err
}

func (s *PostgresStore) RecordLoginSuccess(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `UPDATE users SET failed_login_count=0, locked_until=NULL, updated_at=CURRENT_TIMESTAMP WHERE id=$1`, userID)
	return err
}

func (s *PostgresStore) CreateRefreshToken(ctx context.Context, user User, hash [sha256.Size]byte, expires time.Time) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO refresh_tokens (organization_id,user_id,token_hash,expires_at) VALUES ($1,$2,$3,$4)`, user.OrganizationID, user.ID, hash[:], expires)
	return err
}

func (s *PostgresStore) RotateRefreshToken(ctx context.Context, oldHash, newHash [sha256.Size]byte, expires time.Time) (User, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback(ctx)
	var tokenID string
	var currentExpiry time.Time
	var revoked *time.Time
	var user User
	err = tx.QueryRow(ctx, `SELECT rt.id,rt.expires_at,rt.revoked_at,u.id,u.organization_id,u.email,u.display_name,u.status,u.password_hash,m.role
        FROM refresh_tokens rt JOIN users u ON u.id=rt.user_id JOIN memberships m ON m.user_id=u.id AND m.organization_id=u.organization_id
        WHERE rt.token_hash=$1 FOR UPDATE`, oldHash[:]).Scan(&tokenID, &currentExpiry, &revoked, &user.ID, &user.OrganizationID, &user.Email, &user.DisplayName, &user.Status, &user.PasswordHash, &user.Role)
	if err != nil || revoked != nil || !currentExpiry.After(time.Now().UTC()) || user.Status != "active" {
		return User{}, ErrInvalidToken
	}
	var replacementID string
	err = tx.QueryRow(ctx, `INSERT INTO refresh_tokens (organization_id,user_id,token_hash,expires_at) VALUES ($1,$2,$3,$4) RETURNING id`, user.OrganizationID, user.ID, newHash[:], expires).Scan(&replacementID)
	if err != nil {
		return User{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=CURRENT_TIMESTAMP,replaced_by=$2 WHERE id=$1`, tokenID, replacementID); err != nil {
		return User{}, err
	}
	return user, tx.Commit(ctx)
}

func (s *PostgresStore) RevokeRefreshToken(ctx context.Context, hash [sha256.Size]byte) error {
	command, err := s.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=COALESCE(revoked_at,CURRENT_TIMESTAMP) WHERE token_hash=$1`, hash[:])
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return ErrInvalidToken
	}
	return nil
}

func (s *PostgresStore) Audit(ctx context.Context, user User, action string) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO audit_entries (organization_id,actor_user_id,action,subject_type,subject_id) VALUES ($1,$2,$3,'user',$2)`, user.OrganizationID, user.ID, action)
	return err
}

var _ Store = (*PostgresStore)(nil)
