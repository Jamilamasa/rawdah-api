package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type RecurringTaskRepo struct {
	db *sqlx.DB
}

func NewRecurringTaskRepo(db *sqlx.DB) *RecurringTaskRepo {
	return &RecurringTaskRepo{db: db}
}

func (r *RecurringTaskRepo) List(ctx context.Context, familyID string) ([]*models.RecurringTask, error) {
	var tasks []*models.RecurringTask
	err := r.db.SelectContext(ctx, &tasks,
		`SELECT id, family_id, title, description, assigned_to, created_by, reward_id, is_active, created_at
		 FROM recurring_tasks WHERE family_id = $1 ORDER BY created_at DESC`,
		familyID,
	)
	return tasks, err
}

func (r *RecurringTaskRepo) Create(ctx context.Context, t *models.RecurringTask) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO recurring_tasks (family_id, title, description, assigned_to, created_by, reward_id)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		t.FamilyID, t.Title, t.Description, t.AssignedTo, t.CreatedBy, t.RewardID,
	).Scan(&t.ID, &t.CreatedAt)
}

func (r *RecurringTaskRepo) GetByID(ctx context.Context, id, familyID string) (*models.RecurringTask, error) {
	var t models.RecurringTask
	err := r.db.GetContext(ctx, &t,
		`SELECT id, family_id, title, description, assigned_to, created_by, reward_id, is_active, created_at
		 FROM recurring_tasks WHERE id = $1 AND family_id = $2`,
		id, familyID,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *RecurringTaskRepo) Delete(ctx context.Context, id, familyID string) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM recurring_tasks WHERE id = $1 AND family_id = $2`,
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

func (r *RecurringTaskRepo) ListAllActive(ctx context.Context) ([]*models.RecurringTask, error) {
	var tasks []*models.RecurringTask
	err := r.db.SelectContext(ctx, &tasks,
		`SELECT id, family_id, title, description, assigned_to, created_by, reward_id, is_active, created_at
		 FROM recurring_tasks WHERE is_active = TRUE ORDER BY family_id, created_at`,
	)
	return tasks, err
}
