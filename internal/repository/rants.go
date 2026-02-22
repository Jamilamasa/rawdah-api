package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type RantRepo struct {
	db *sqlx.DB
}

func NewRantRepo(db *sqlx.DB) *RantRepo {
	return &RantRepo{db: db}
}

func (r *RantRepo) Create(ctx context.Context, rant *models.Rant) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO rants (user_id, title, content, password_hash)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		rant.UserID, rant.Title, rant.Content, rant.PasswordHash,
	).Scan(&rant.ID, &rant.CreatedAt)
}

// List returns rants for a user. Content is included only if not locked.
func (r *RantRepo) List(ctx context.Context, userID string) ([]*models.Rant, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, title, content, password_hash, created_at FROM rants WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rants []*models.Rant
	for rows.Next() {
		var rant models.Rant
		if err := rows.Scan(&rant.ID, &rant.UserID, &rant.Title, &rant.Content, &rant.PasswordHash, &rant.CreatedAt); err != nil {
			return nil, err
		}
		rant.IsLocked = rant.PasswordHash != nil
		// Omit content if locked (it will be served as empty in JSON via omitempty)
		if rant.IsLocked {
			rant.Content = ""
		}
		rants = append(rants, &rant)
	}
	return rants, rows.Err()
}

// GetByID returns a rant with its password_hash for validation.
func (r *RantRepo) GetByID(ctx context.Context, id, userID string) (*models.Rant, error) {
	var rant models.Rant
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, title, content, password_hash, created_at FROM rants WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&rant.ID, &rant.UserID, &rant.Title, &rant.Content, &rant.PasswordHash, &rant.CreatedAt)
	if err != nil {
		return nil, err
	}
	rant.IsLocked = rant.PasswordHash != nil
	return &rant, nil
}

func (r *RantRepo) Update(ctx context.Context, id, userID string, title *string, content string, passwordHash *string) (*models.Rant, error) {
	var rant models.Rant
	err := r.db.QueryRowContext(ctx,
		`UPDATE rants SET title = COALESCE($1, title), content = $2, password_hash = $3
		 WHERE id = $4 AND user_id = $5
		 RETURNING id, user_id, title, content, password_hash, created_at`,
		title, content, passwordHash, id, userID,
	).Scan(&rant.ID, &rant.UserID, &rant.Title, &rant.Content, &rant.PasswordHash, &rant.CreatedAt)
	if err != nil {
		return nil, err
	}
	rant.IsLocked = rant.PasswordHash != nil
	return &rant, nil
}

func (r *RantRepo) Delete(ctx context.Context, id, userID string) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM rants WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}
