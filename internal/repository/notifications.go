package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type NotificationRepo struct {
	db *sqlx.DB
}

func NewNotificationRepo(db *sqlx.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Create(ctx context.Context, notif *models.Notification) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO notifications (user_id, type, title, body, data)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		notif.UserID, notif.Type, notif.Title, notif.Body, notif.Data,
	).Scan(&notif.ID, &notif.CreatedAt)
}

func (r *NotificationRepo) List(ctx context.Context, userID string) ([]*models.Notification, error) {
	var notifs []*models.Notification
	err := r.db.SelectContext(ctx, &notifs,
		`SELECT id, user_id, type, title, body, data, is_read, created_at
		 FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT 100`,
		userID,
	)
	return notifs, err
}

func (r *NotificationRepo) ReadOne(ctx context.Context, id, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	return err
}

func (r *NotificationRepo) ReadAll(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE`,
		userID,
	)
	return err
}
