package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetByID(ctx context.Context, id, familyID string) (*models.User, error) {
	var u models.User
	err := r.db.GetContext(ctx, &u,
		`SELECT id, family_id, role, name, username, email, password_hash, avatar_url,
		        theme, date_of_birth, game_limit_minutes, is_active, created_by, last_login_at, created_at
		 FROM users WHERE id = $1 AND family_id = $2`,
		id, familyID,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.db.GetContext(ctx, &u,
		`SELECT id, family_id, role, name, username, email, password_hash, avatar_url,
		        theme, date_of_birth, game_limit_minutes, is_active, created_by, last_login_at, created_at
		 FROM users WHERE email = $1`,
		email,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByUsernameAndFamily(ctx context.Context, username, familyID string) (*models.User, error) {
	var u models.User
	err := r.db.GetContext(ctx, &u,
		`SELECT id, family_id, role, name, username, email, password_hash, avatar_url,
		        theme, date_of_birth, game_limit_minutes, is_active, created_by, last_login_at, created_at
		 FROM users WHERE username = $1 AND family_id = $2`,
		username, familyID,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET last_login_at = $1 WHERE id = $2`,
		now, userID,
	)
	return err
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, hash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $1 WHERE id = $2`,
		hash, userID,
	)
	return err
}

func (r *UserRepo) UpdateAvatar(ctx context.Context, userID uuid.UUID, familyID, url string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET avatar_url = $1 WHERE id = $2 AND family_id = $3`,
		url, userID, familyID,
	)
	return err
}
