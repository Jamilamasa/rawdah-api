package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type LessonRepo struct {
	db *sqlx.DB
}

func NewLessonRepo(db *sqlx.DB) *LessonRepo {
	return &LessonRepo{db: db}
}

// ---- Quran Lessons ----

func (r *LessonRepo) CreateLesson(ctx context.Context, lesson *models.QuranLesson) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO quran_lessons (family_id, verse_id, assigned_to, assigned_by, reward_id, status)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		lesson.FamilyID, lesson.VerseID, lesson.AssignedTo, lesson.AssignedBy, lesson.RewardID, lesson.Status,
	).Scan(&lesson.ID, &lesson.CreatedAt)
}

func (r *LessonRepo) GetLesson(ctx context.Context, id, familyID string) (*models.QuranLesson, error) {
	var l models.QuranLesson
	err := r.db.GetContext(ctx, &l,
		`SELECT id, family_id, verse_id, assigned_to, assigned_by, reward_id, status, completed_at, created_at
		 FROM quran_lessons WHERE id = $1 AND family_id = $2`,
		id, familyID,
	)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *LessonRepo) ListLessons(ctx context.Context, familyID string) ([]*models.QuranLesson, error) {
	var lessons []*models.QuranLesson
	err := r.db.SelectContext(ctx, &lessons,
		`SELECT id, family_id, verse_id, assigned_to, assigned_by, reward_id, status, completed_at, created_at
		 FROM quran_lessons WHERE family_id = $1 ORDER BY created_at DESC`,
		familyID,
	)
	return lessons, err
}

func (r *LessonRepo) ListMyLessons(ctx context.Context, userID, familyID string) ([]*models.QuranLesson, error) {
	var lessons []*models.QuranLesson
	err := r.db.SelectContext(ctx, &lessons,
		`SELECT id, family_id, verse_id, assigned_to, assigned_by, reward_id, status, completed_at, created_at
		 FROM quran_lessons WHERE assigned_to = $1 AND family_id = $2 ORDER BY created_at DESC`,
		userID, familyID,
	)
	return lessons, err
}

func (r *LessonRepo) CompleteLesson(ctx context.Context, id uuid.UUID, userID, familyID string) error {
	now := time.Now()
	result, err := r.db.ExecContext(ctx,
		`UPDATE quran_lessons SET status = 'completed', completed_at = $1
		 WHERE id = $2 AND assigned_to = $3 AND family_id = $4 AND status != 'completed'`,
		now, id, userID, familyID,
	)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return nil // already completed or not found
	}
	return nil
}

func (r *LessonRepo) CountCompletedLessons(ctx context.Context, userID, familyID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM quran_lessons
		 WHERE assigned_to = $1 AND family_id = $2 AND status = 'completed'`,
		userID, familyID,
	).Scan(&count)
	return count, err
}

// ---- Learn Content ----

func (r *LessonRepo) GetLearnContent(ctx context.Context, familyID string) ([]*models.LearnContent, error) {
	var items []*models.LearnContent
	err := r.db.SelectContext(ctx, &items,
		`SELECT id, family_id, assigned_to, title, content_type, content, reward_id, created_by, created_at
		 FROM learn_content WHERE family_id = $1 ORDER BY created_at DESC`,
		familyID,
	)
	return items, err
}

func (r *LessonRepo) GetMyLearnContent(ctx context.Context, userID, familyID string) ([]*models.LearnContent, error) {
	var items []*models.LearnContent
	err := r.db.SelectContext(ctx, &items,
		`SELECT id, family_id, assigned_to, title, content_type, content, reward_id, created_by, created_at
		 FROM learn_content WHERE family_id = $1 AND (assigned_to = $2 OR assigned_to IS NULL) ORDER BY created_at DESC`,
		familyID, userID,
	)
	return items, err
}

func (r *LessonRepo) CreateLearnContent(ctx context.Context, content *models.LearnContent) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO learn_content (family_id, assigned_to, title, content_type, content, reward_id, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at`,
		content.FamilyID, content.AssignedTo, content.Title, content.ContentType,
		content.Content, content.RewardID, content.CreatedBy,
	).Scan(&content.ID, &content.CreatedAt)
}

func (r *LessonRepo) CompleteLearnContent(ctx context.Context, contentID uuid.UUID, userID, familyID string) error {
	// Verify content belongs to family
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM learn_content WHERE id = $1 AND family_id = $2)`,
		contentID, familyID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	now := time.Now()
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO learn_progress (content_id, user_id, completed_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (content_id, user_id) DO UPDATE SET completed_at = $3`,
		contentID, userID, now,
	)
	return err
}

func (r *LessonRepo) GetLearnProgress(ctx context.Context, userID, familyID string) ([]*models.LearnProgress, error) {
	var progress []*models.LearnProgress
	err := r.db.SelectContext(ctx, &progress,
		`SELECT lp.id, lp.content_id, lp.user_id, lp.completed_at
		 FROM learn_progress lp
		 JOIN learn_content lc ON lc.id = lp.content_id
		 WHERE lp.user_id = $1 AND lc.family_id = $2`,
		userID, familyID,
	)
	return progress, err
}
