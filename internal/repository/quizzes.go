package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type QuizRepo struct {
	db *sqlx.DB
}

func NewQuizRepo(db *sqlx.DB) *QuizRepo {
	return &QuizRepo{db: db}
}

// ---- Hadith Quizzes ----

func (r *QuizRepo) CreateHadithQuiz(ctx context.Context, q *models.HadithQuiz) error {
	questionsJSON, err := json.Marshal(q.Questions)
	if err != nil {
		return err
	}
	return r.db.QueryRowContext(ctx,
		`INSERT INTO hadith_quizzes (family_id, hadith_id, assigned_to, assigned_by, questions, status, memorize_until)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at`,
		q.FamilyID, q.HadithID, q.AssignedTo, q.AssignedBy, questionsJSON, q.Status, q.MemorizeUntil,
	).Scan(&q.ID, &q.CreatedAt)
}

func (r *QuizRepo) GetHadithQuizByID(ctx context.Context, id, familyID string) (*models.HadithQuiz, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, family_id, hadith_id, assigned_to, assigned_by, questions, answers, score,
		        xp_awarded, status, memorize_until, completed_at, created_at
		 FROM hadith_quizzes WHERE id = $1 AND family_id = $2`,
		id, familyID,
	)
	return scanHadithQuiz(row)
}

func scanHadithQuiz(row interface{ Scan(...interface{}) error }) (*models.HadithQuiz, error) {
	var q models.HadithQuiz
	var questionsRaw, answersRaw []byte
	err := row.Scan(
		&q.ID, &q.FamilyID, &q.HadithID, &q.AssignedTo, &q.AssignedBy,
		&questionsRaw, &answersRaw, &q.Score, &q.XPAwarded, &q.Status,
		&q.MemorizeUntil, &q.CompletedAt, &q.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(questionsRaw, &q.Questions); err != nil {
		return nil, err
	}
	if answersRaw != nil {
		if err := json.Unmarshal(answersRaw, &q.Answers); err != nil {
			return nil, err
		}
	}
	return &q, nil
}

func (r *QuizRepo) ListHadithQuizzesByFamily(ctx context.Context, familyID string) ([]*models.HadithQuiz, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, family_id, hadith_id, assigned_to, assigned_by, questions, answers, score,
		        xp_awarded, status, memorize_until, completed_at, created_at
		 FROM hadith_quizzes WHERE family_id = $1 ORDER BY created_at DESC`,
		familyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*models.HadithQuiz
	for rows.Next() {
		q, err := scanHadithQuiz(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, q)
	}
	return result, rows.Err()
}

func (r *QuizRepo) ListMyHadithQuizzes(ctx context.Context, userID, familyID string) ([]*models.HadithQuiz, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, family_id, hadith_id, assigned_to, assigned_by, questions, answers, score,
		        xp_awarded, status, memorize_until, completed_at, created_at
		 FROM hadith_quizzes WHERE assigned_to = $1 AND family_id = $2 ORDER BY created_at DESC`,
		userID, familyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*models.HadithQuiz
	for rows.Next() {
		q, err := scanHadithQuiz(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, q)
	}
	return result, rows.Err()
}

func (r *QuizRepo) UpdateHadithQuizStatus(ctx context.Context, id, familyID, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE hadith_quizzes SET status = $1 WHERE id = $2 AND family_id = $3`,
		status, id, familyID,
	)
	return err
}

func (r *QuizRepo) SubmitHadithQuiz(ctx context.Context, id, familyID string, answers []models.QuizAnswer, score, xp int) error {
	answersJSON, err := json.Marshal(answers)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE hadith_quizzes SET answers = $1, score = $2, xp_awarded = $3, status = 'completed', completed_at = NOW()
		 WHERE id = $4 AND family_id = $5`,
		answersJSON, score, xp, id, familyID,
	)
	return err
}

// ---- Prophet Quizzes ----

func (r *QuizRepo) CreateProphetQuiz(ctx context.Context, q *models.ProphetQuiz) error {
	questionsJSON, err := json.Marshal(q.Questions)
	if err != nil {
		return err
	}
	return r.db.QueryRowContext(ctx,
		`INSERT INTO prophet_quizzes (family_id, prophet_id, assigned_to, assigned_by, questions, status)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		q.FamilyID, q.ProphetID, q.AssignedTo, q.AssignedBy, questionsJSON, q.Status,
	).Scan(&q.ID, &q.CreatedAt)
}

func (r *QuizRepo) GetProphetQuizByID(ctx context.Context, id, familyID string) (*models.ProphetQuiz, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, family_id, prophet_id, assigned_to, assigned_by, questions, answers, score,
		        xp_awarded, status, completed_at, created_at
		 FROM prophet_quizzes WHERE id = $1 AND family_id = $2`,
		id, familyID,
	)
	return scanProphetQuiz(row)
}

func scanProphetQuiz(row interface{ Scan(...interface{}) error }) (*models.ProphetQuiz, error) {
	var q models.ProphetQuiz
	var questionsRaw, answersRaw []byte
	err := row.Scan(
		&q.ID, &q.FamilyID, &q.ProphetID, &q.AssignedTo, &q.AssignedBy,
		&questionsRaw, &answersRaw, &q.Score, &q.XPAwarded, &q.Status,
		&q.CompletedAt, &q.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(questionsRaw, &q.Questions); err != nil {
		return nil, err
	}
	if answersRaw != nil {
		if err := json.Unmarshal(answersRaw, &q.Answers); err != nil {
			return nil, err
		}
	}
	return &q, nil
}

func (r *QuizRepo) ListProphetQuizzesByFamily(ctx context.Context, familyID string) ([]*models.ProphetQuiz, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, family_id, prophet_id, assigned_to, assigned_by, questions, answers, score,
		        xp_awarded, status, completed_at, created_at
		 FROM prophet_quizzes WHERE family_id = $1 ORDER BY created_at DESC`,
		familyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*models.ProphetQuiz
	for rows.Next() {
		q, err := scanProphetQuiz(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, q)
	}
	return result, rows.Err()
}

func (r *QuizRepo) ListMyProphetQuizzes(ctx context.Context, userID, familyID string) ([]*models.ProphetQuiz, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, family_id, prophet_id, assigned_to, assigned_by, questions, answers, score,
		        xp_awarded, status, completed_at, created_at
		 FROM prophet_quizzes WHERE assigned_to = $1 AND family_id = $2 ORDER BY created_at DESC`,
		userID, familyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*models.ProphetQuiz
	for rows.Next() {
		q, err := scanProphetQuiz(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, q)
	}
	return result, rows.Err()
}

func (r *QuizRepo) UpdateProphetQuizStatus(ctx context.Context, id, familyID, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE prophet_quizzes SET status = $1 WHERE id = $2 AND family_id = $3`,
		status, id, familyID,
	)
	return err
}

func (r *QuizRepo) SubmitProphetQuiz(ctx context.Context, id, familyID string, answers []models.QuizAnswer, score, xp int) error {
	answersJSON, err := json.Marshal(answers)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE prophet_quizzes SET answers = $1, score = $2, xp_awarded = $3, status = 'completed', completed_at = NOW()
		 WHERE id = $4 AND family_id = $5`,
		answersJSON, score, xp, id, familyID,
	)
	return err
}

// ---- Quran Quizzes ----

func (r *QuizRepo) CreateQuranQuiz(ctx context.Context, q *models.QuranQuiz) error {
	questionsJSON, err := json.Marshal(q.Questions)
	if err != nil {
		return err
	}
	return r.db.QueryRowContext(ctx,
		`INSERT INTO quran_quizzes (family_id, verse_id, lesson_id, assigned_to, questions, status)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		q.FamilyID, q.VerseID, q.LessonID, q.AssignedTo, questionsJSON, q.Status,
	).Scan(&q.ID, &q.CreatedAt)
}

func (r *QuizRepo) GetQuranQuizByID(ctx context.Context, id, familyID string) (*models.QuranQuiz, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, family_id, verse_id, lesson_id, assigned_to, questions, answers, score,
		        xp_awarded, status, completed_at, created_at
		 FROM quran_quizzes WHERE id = $1 AND family_id = $2`,
		id, familyID,
	)
	return scanQuranQuiz(row)
}

func scanQuranQuiz(row interface{ Scan(...interface{}) error }) (*models.QuranQuiz, error) {
	var q models.QuranQuiz
	var questionsRaw, answersRaw []byte
	err := row.Scan(
		&q.ID, &q.FamilyID, &q.VerseID, &q.LessonID, &q.AssignedTo,
		&questionsRaw, &answersRaw, &q.Score, &q.XPAwarded, &q.Status,
		&q.CompletedAt, &q.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(questionsRaw, &q.Questions); err != nil {
		return nil, err
	}
	if answersRaw != nil {
		if err := json.Unmarshal(answersRaw, &q.Answers); err != nil {
			return nil, err
		}
	}
	return &q, nil
}

func (r *QuizRepo) ListQuranQuizzesByFamily(ctx context.Context, familyID string) ([]*models.QuranQuiz, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, family_id, verse_id, lesson_id, assigned_to, questions, answers, score,
		        xp_awarded, status, completed_at, created_at
		 FROM quran_quizzes WHERE family_id = $1 ORDER BY created_at DESC`,
		familyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*models.QuranQuiz
	for rows.Next() {
		q, err := scanQuranQuiz(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, q)
	}
	return result, rows.Err()
}

func (r *QuizRepo) ListMyQuranQuizzes(ctx context.Context, userID, familyID string) ([]*models.QuranQuiz, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, family_id, verse_id, lesson_id, assigned_to, questions, answers, score,
		        xp_awarded, status, completed_at, created_at
		 FROM quran_quizzes WHERE assigned_to = $1 AND family_id = $2 ORDER BY created_at DESC`,
		userID, familyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*models.QuranQuiz
	for rows.Next() {
		q, err := scanQuranQuiz(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, q)
	}
	return result, rows.Err()
}

func (r *QuizRepo) UpdateQuranQuizStatus(ctx context.Context, id, familyID, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE quran_quizzes SET status = $1 WHERE id = $2 AND family_id = $3`,
		status, id, familyID,
	)
	return err
}

func (r *QuizRepo) SubmitQuranQuiz(ctx context.Context, id, familyID string, answers []models.QuizAnswer, score, xp int) error {
	answersJSON, err := json.Marshal(answers)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE quran_quizzes SET answers = $1, score = $2, xp_awarded = $3, status = 'completed', completed_at = NOW()
		 WHERE id = $4 AND family_id = $5`,
		answersJSON, score, xp, id, familyID,
	)
	return err
}

// CountCompletedHadithQuizzes counts completed hadith quizzes for a user
func (r *QuizRepo) CountCompletedHadithQuizzes(ctx context.Context, userID, familyID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM hadith_quizzes
		 WHERE assigned_to = $1 AND family_id = $2 AND status = 'completed'`,
		userID, familyID,
	).Scan(&count)
	return count, err
}

// CountDistinctCompletedProphetQuizzes counts distinct prophets with completed quizzes
func (r *QuizRepo) CountDistinctCompletedProphetQuizzes(ctx context.Context, userID, familyID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT prophet_id) FROM prophet_quizzes
		 WHERE assigned_to = $1 AND family_id = $2 AND status = 'completed'`,
		userID, familyID,
	).Scan(&count)
	return count, err
}

// HasPerfectScoreQuiz returns true if the user has any 100% quiz
func (r *QuizRepo) HasPerfectScoreQuiz(ctx context.Context, userID, familyID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
		   SELECT 1 FROM hadith_quizzes WHERE assigned_to = $1 AND family_id = $2 AND score = 100
		   UNION ALL
		   SELECT 1 FROM prophet_quizzes WHERE assigned_to = $1 AND family_id = $2 AND score = 100
		   UNION ALL
		   SELECT 1 FROM quran_quizzes WHERE assigned_to = $1 AND family_id = $2 AND score = 100
		 )`,
		userID, familyID,
	).Scan(&exists)
	return exists, err
}

// CountAllCompletedQuizzes counts total completed quizzes across all types
func (r *QuizRepo) CountAllCompletedQuizzes(ctx context.Context, userID, familyID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT (
		   SELECT COUNT(*) FROM hadith_quizzes
		   WHERE assigned_to = $1 AND family_id = $2 AND status = 'completed'
		 ) + (
		   SELECT COUNT(*) FROM prophet_quizzes
		   WHERE assigned_to = $1 AND family_id = $2 AND status = 'completed'
		 ) + (
		   SELECT COUNT(*) FROM quran_quizzes
		   WHERE assigned_to = $1 AND family_id = $2 AND status = 'completed'
		 ) AS total`,
		userID, familyID,
	).Scan(&count)
	return count, err
}

// GetQuranQuizByLessonID retrieves a quran quiz by lesson_id and family_id
func (r *QuizRepo) GetQuranQuizByLessonID(ctx context.Context, lessonID uuid.UUID, familyID string) (*models.QuranQuiz, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, family_id, verse_id, lesson_id, assigned_to, questions, answers, score,
		        xp_awarded, status, completed_at, created_at
		 FROM quran_quizzes WHERE lesson_id = $1 AND family_id = $2 LIMIT 1`,
		lessonID, familyID,
	)
	q, err := scanQuranQuiz(row)
	if err != nil {
		return nil, fmt.Errorf("quiz not found: %w", err)
	}
	return q, nil
}
