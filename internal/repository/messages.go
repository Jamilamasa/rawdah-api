package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type MessageRepo struct {
	db *sqlx.DB
}

func NewMessageRepo(db *sqlx.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

func (r *MessageRepo) Send(ctx context.Context, msg *models.Message) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO messages (family_id, sender_id, recipient_id, content)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		msg.FamilyID, msg.SenderID, msg.RecipientID, msg.Content,
	).Scan(&msg.ID, &msg.CreatedAt)
}

func (r *MessageRepo) GetThread(ctx context.Context, userID, otherUserID, familyID string) ([]*models.Message, error) {
	var messages []*models.Message
	err := r.db.SelectContext(ctx, &messages,
		`SELECT id, family_id, sender_id, recipient_id, content, read_at, created_at
		 FROM messages
		 WHERE family_id = $1
		   AND ((sender_id = $2 AND recipient_id = $3) OR (sender_id = $3 AND recipient_id = $2))
		 ORDER BY created_at ASC`,
		familyID, userID, otherUserID,
	)
	return messages, err
}

func (r *MessageRepo) Conversations(ctx context.Context, userID, familyID string) ([]*models.Message, error) {
	var messages []*models.Message
	err := r.db.SelectContext(ctx, &messages,
		`SELECT DISTINCT ON (LEAST(sender_id::text, recipient_id::text), GREATEST(sender_id::text, recipient_id::text))
		        id, family_id, sender_id, recipient_id, content, read_at, created_at
		 FROM messages
		 WHERE family_id = $1 AND (sender_id = $2 OR recipient_id = $2)
		 ORDER BY LEAST(sender_id::text, recipient_id::text), GREATEST(sender_id::text, recipient_id::text), created_at DESC`,
		familyID, userID,
	)
	return messages, err
}

func (r *MessageRepo) MarkRead(ctx context.Context, id, recipientID, familyID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE messages SET read_at = NOW()
		 WHERE id = $1 AND recipient_id = $2 AND family_id = $3 AND read_at IS NULL`,
		id, recipientID, familyID,
	)
	return err
}

func (r *MessageRepo) CountSent(ctx context.Context, userID, familyID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM messages WHERE sender_id = $1 AND family_id = $2`,
		userID, familyID,
	).Scan(&count)
	return count, err
}
