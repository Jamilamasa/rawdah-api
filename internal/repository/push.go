package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type PushRepo struct {
	db *sqlx.DB
}

func NewPushRepo(db *sqlx.DB) *PushRepo {
	return &PushRepo{db: db}
}

func (r *PushRepo) Subscribe(ctx context.Context, sub *models.PushSubscription) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (endpoint) DO UPDATE SET user_id = $1, p256dh = $3, auth = $4
		 RETURNING id, created_at`,
		sub.UserID, sub.Endpoint, sub.P256dh, sub.Auth,
	).Scan(&sub.ID, &sub.CreatedAt)
}

func (r *PushRepo) Unsubscribe(ctx context.Context, endpoint, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM push_subscriptions WHERE endpoint = $1 AND user_id = $2`,
		endpoint, userID,
	)
	return err
}

func (r *PushRepo) GetSubscriptions(ctx context.Context, userID string) ([]*models.PushSubscription, error) {
	var subs []*models.PushSubscription
	err := r.db.SelectContext(ctx, &subs,
		`SELECT id, user_id, endpoint, p256dh, auth, created_at
		 FROM push_subscriptions WHERE user_id = $1`,
		userID,
	)
	return subs, err
}
