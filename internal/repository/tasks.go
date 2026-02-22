package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type TaskRepo struct {
	db *sqlx.DB
}

func NewTaskRepo(db *sqlx.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

type TaskFilter struct {
	Status     string
	AssignedTo string
}

type DueRewardFilter struct {
	Status     string
	AssignedTo string
}

func (r *TaskRepo) List(ctx context.Context, familyID string, filter TaskFilter) ([]*models.Task, error) {
	query := `SELECT id, family_id, title, description, assigned_to, created_by, reward_id,
	                 status, due_date, completed_at, created_at
	          FROM tasks WHERE family_id = $1`
	args := []interface{}{familyID}
	argIdx := 2

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.AssignedTo != "" {
		query += fmt.Sprintf(" AND assigned_to = $%d", argIdx)
		args = append(args, filter.AssignedTo)
		argIdx++
	}
	query += " ORDER BY created_at DESC"

	var tasks []*models.Task
	err := r.db.SelectContext(ctx, &tasks, query, args...)
	return tasks, err
}

func (r *TaskRepo) ListDueRewards(ctx context.Context, familyID string, filter DueRewardFilter) ([]*models.DueReward, error) {
	query := `SELECT
		t.id AS task_id,
		t.title AS task_title,
		t.description AS task_description,
		t.status AS task_status,
		t.due_date AS task_due_date,
		t.completed_at AS task_completed_at,
		t.created_at AS task_created_at,
		u.id AS child_id,
		u.name AS child_name,
		r.id AS reward_id,
		r.title AS reward_title,
		r.description AS reward_description,
		r.value AS reward_value,
		r.type AS reward_type,
		r.icon AS reward_icon
	FROM tasks t
	JOIN users u ON u.id = t.assigned_to AND u.family_id = t.family_id
	JOIN rewards r ON r.id = t.reward_id AND r.family_id = t.family_id
	WHERE t.family_id = $1
	  AND t.status IN ('reward_requested', 'reward_approved')`
	args := []interface{}{familyID}
	argIdx := 2

	if filter.Status != "" {
		query += fmt.Sprintf(" AND t.status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.AssignedTo != "" {
		query += fmt.Sprintf(" AND t.assigned_to = $%d", argIdx)
		args = append(args, filter.AssignedTo)
		argIdx++
	}

	query += " ORDER BY t.completed_at DESC NULLS LAST, t.created_at DESC"

	var dueRewards []*models.DueReward
	if err := r.db.SelectContext(ctx, &dueRewards, query, args...); err != nil {
		return nil, err
	}
	return dueRewards, nil
}

func (r *TaskRepo) Create(ctx context.Context, task *models.Task) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO tasks (family_id, title, description, assigned_to, created_by, reward_id, status, due_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, created_at`,
		task.FamilyID, task.Title, task.Description, task.AssignedTo, task.CreatedBy,
		task.RewardID, task.Status, task.DueDate,
	).Scan(&task.ID, &task.CreatedAt)
}

func (r *TaskRepo) GetByID(ctx context.Context, id, familyID string) (*models.Task, error) {
	var t models.Task
	err := r.db.GetContext(ctx, &t,
		`SELECT id, family_id, title, description, assigned_to, created_by, reward_id,
		        status, due_date, completed_at, created_at
		 FROM tasks WHERE id = $1 AND family_id = $2`,
		id, familyID,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TaskRepo) UpdateStatus(ctx context.Context, id, familyID, status string) (*models.Task, error) {
	var completedAt *time.Time
	if status == "completed" {
		now := time.Now()
		completedAt = &now
	}

	var t models.Task
	err := r.db.GetContext(ctx, &t,
		`UPDATE tasks SET status = $1, completed_at = COALESCE($2, completed_at)
		 WHERE id = $3 AND family_id = $4
		 RETURNING id, family_id, title, description, assigned_to, created_by, reward_id,
		           status, due_date, completed_at, created_at`,
		status, completedAt, id, familyID,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TaskRepo) Update(ctx context.Context, id, familyID string, title string, description *string, rewardID *uuid.UUID, dueDate *time.Time) (*models.Task, error) {
	var t models.Task
	err := r.db.GetContext(ctx, &t,
		`UPDATE tasks SET title = $1, description = $2, reward_id = $3, due_date = $4
		 WHERE id = $5 AND family_id = $6
		 RETURNING id, family_id, title, description, assigned_to, created_by, reward_id,
		           status, due_date, completed_at, created_at`,
		title, description, rewardID, dueDate, id, familyID,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TaskRepo) Delete(ctx context.Context, id, familyID string) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM tasks WHERE id = $1 AND family_id = $2`,
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

func (r *TaskRepo) CountCompleted(ctx context.Context, userID, familyID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tasks WHERE assigned_to = $1 AND family_id = $2 AND status IN ('completed', 'reward_requested', 'reward_approved')`,
		userID, familyID,
	).Scan(&count)
	return count, err
}
