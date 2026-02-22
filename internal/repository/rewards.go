package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type RewardRepo struct {
	db *sqlx.DB
}

func NewRewardRepo(db *sqlx.DB) *RewardRepo {
	return &RewardRepo{db: db}
}

func (r *RewardRepo) List(ctx context.Context, familyID string) ([]*models.Reward, error) {
	var rewards []*models.Reward
	err := r.db.SelectContext(ctx, &rewards,
		`SELECT id, family_id, title, description, value, type, icon, created_by, created_at
		 FROM rewards WHERE family_id = $1 ORDER BY created_at DESC`,
		familyID,
	)
	return rewards, err
}

func (r *RewardRepo) Create(ctx context.Context, reward *models.Reward) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO rewards (family_id, title, description, value, type, icon, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at`,
		reward.FamilyID, reward.Title, reward.Description, reward.Value, reward.Type, reward.Icon, reward.CreatedBy,
	).Scan(&reward.ID, &reward.CreatedAt)
}

func (r *RewardRepo) GetByID(ctx context.Context, id, familyID string) (*models.Reward, error) {
	var rw models.Reward
	err := r.db.GetContext(ctx, &rw,
		`SELECT id, family_id, title, description, value, type, icon, created_by, created_at
		 FROM rewards WHERE id = $1 AND family_id = $2`,
		id, familyID,
	)
	if err != nil {
		return nil, err
	}
	return &rw, nil
}

func (r *RewardRepo) Update(ctx context.Context, id, familyID string, title string, description *string, value float64, rewardType string, icon *string) (*models.Reward, error) {
	var rw models.Reward
	err := r.db.GetContext(ctx, &rw,
		`UPDATE rewards SET title = $1, description = $2, value = $3, type = $4, icon = $5
		 WHERE id = $6 AND family_id = $7
		 RETURNING id, family_id, title, description, value, type, icon, created_by, created_at`,
		title, description, value, rewardType, icon, id, familyID,
	)
	if err != nil {
		return nil, err
	}
	return &rw, nil
}

func (r *RewardRepo) Delete(ctx context.Context, id, familyID string) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM rewards WHERE id = $1 AND family_id = $2`,
		id, familyID,
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
