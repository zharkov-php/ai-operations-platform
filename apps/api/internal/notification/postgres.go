package notification

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Item struct {
	ID        string    `json:"id"`
	EventType string    `json:"event_type"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	DeepLink  string    `json:"deep_link"`
	Status    string    `json:"status"`
	Attempts  int       `json:"attempts"`
	CreatedAt time.Time `json:"created_at"`
}
type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }
func (s *Store) Publish(ctx context.Context, orgID, userID string, event Event) (Item, bool, error) {
	if err := Validate(event); err != nil {
		return Item{}, false, err
	}
	var item Item
	err := s.pool.QueryRow(ctx, `INSERT INTO notifications(organization_id,user_id,event_type,title,body,deep_link,dedupe_key,status,delivered_at) VALUES($1,$2,$3,$4,$5,$6,$7,'delivered',CURRENT_TIMESTAMP) ON CONFLICT(user_id,dedupe_key) DO NOTHING RETURNING id,event_type,title,body,deep_link,status,attempts,created_at`, orgID, userID, event.Type, event.Title, event.Body, event.DeepLink, event.DedupeKey).Scan(&item.ID, &item.EventType, &item.Title, &item.Body, &item.DeepLink, &item.Status, &item.Attempts, &item.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, false, nil
	}
	return item, err == nil, err
}
func (s *Store) List(ctx context.Context, orgID, userID string) ([]Item, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,event_type,title,body,deep_link,status,attempts,created_at FROM notifications WHERE organization_id=$1 AND user_id=$2 ORDER BY created_at DESC LIMIT 100`, orgID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Item{}
	for rows.Next() {
		var x Item
		if err = rows.Scan(&x.ID, &x.EventType, &x.Title, &x.Body, &x.DeepLink, &x.Status, &x.Attempts, &x.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
func (s *Store) GetPreferences(ctx context.Context, userID string) (Preferences, error) {
	var p Preferences
	err := s.pool.QueryRow(ctx, `SELECT in_app,mobile_push FROM notification_preferences WHERE user_id=$1`, userID).Scan(&p.InApp, &p.MobilePush)
	if errors.Is(err, pgx.ErrNoRows) {
		return Preferences{InApp: true}, nil
	}
	return p, err
}
func (s *Store) SetPreferences(ctx context.Context, userID string, p Preferences) (Preferences, error) {
	err := s.pool.QueryRow(ctx, `INSERT INTO notification_preferences(user_id,in_app,mobile_push) VALUES($1,$2,$3) ON CONFLICT(user_id) DO UPDATE SET in_app=$2,mobile_push=$3,updated_at=CURRENT_TIMESTAMP RETURNING in_app,mobile_push`, userID, p.InApp, p.MobilePush).Scan(&p.InApp, &p.MobilePush)
	return p, err
}
