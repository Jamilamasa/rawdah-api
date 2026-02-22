package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type DashboardRepo struct {
	db *sqlx.DB
}

func NewDashboardRepo(db *sqlx.DB) *DashboardRepo {
	return &DashboardRepo{db: db}
}

type DashboardSummary struct {
	TotalMembers     int `json:"total_members"`
	ActiveTasks      int `json:"active_tasks"`
	CompletedTasks   int `json:"completed_tasks"`
	PendingRequests  int `json:"pending_requests"`
	TotalGameMinutes int `json:"total_game_minutes"`
	QuizzesCompleted int `json:"quizzes_completed"`
}

func (r *DashboardRepo) Summary(ctx context.Context, familyID string) (*DashboardSummary, error) {
	s := &DashboardSummary{}
	err := r.db.QueryRowContext(ctx,
		`SELECT
		   (SELECT COUNT(*) FROM users WHERE family_id = $1 AND is_active = TRUE) AS total_members,
		   (SELECT COUNT(*) FROM tasks WHERE family_id = $1 AND status IN ('pending', 'in_progress')) AS active_tasks,
		   (SELECT COUNT(*) FROM tasks WHERE family_id = $1 AND status IN ('completed', 'reward_requested', 'reward_approved')) AS completed_tasks,
		   (SELECT COUNT(*) FROM requests WHERE family_id = $1 AND status = 'pending') AS pending_requests,
		   (SELECT COALESCE(SUM(duration_seconds) / 60, 0) FROM game_sessions WHERE family_id = $1) AS total_game_minutes,
		   (SELECT (
		     (SELECT COUNT(*) FROM hadith_quizzes WHERE family_id = $1 AND status = 'completed') +
		     (SELECT COUNT(*) FROM prophet_quizzes WHERE family_id = $1 AND status = 'completed') +
		     (SELECT COUNT(*) FROM quran_quizzes WHERE family_id = $1 AND status = 'completed')
		   )) AS quizzes_completed`,
		familyID,
	).Scan(&s.TotalMembers, &s.ActiveTasks, &s.CompletedTasks, &s.PendingRequests, &s.TotalGameMinutes, &s.QuizzesCompleted)
	if err != nil {
		return nil, err
	}
	return s, nil
}

type DailyTaskCompletion struct {
	Date      string `db:"day"   json:"date"`
	Completed int    `db:"count" json:"completed"`
}

func (r *DashboardRepo) TaskCompletion(ctx context.Context, familyID string, days int) ([]*DailyTaskCompletion, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DATE(completed_at)::TEXT AS day, COUNT(*) AS count
		 FROM tasks
		 WHERE family_id = $1 AND completed_at >= NOW() - ($2 || ' days')::INTERVAL
		 GROUP BY day ORDER BY day ASC`,
		familyID, days,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*DailyTaskCompletion
	for rows.Next() {
		var d DailyTaskCompletion
		if err := rows.Scan(&d.Date, &d.Completed); err != nil {
			return nil, err
		}
		result = append(result, &d)
	}
	return result, rows.Err()
}

type DailyGameTime struct {
	Date    string `json:"date"`
	Minutes int    `json:"minutes"`
}

func (r *DashboardRepo) GameTime(ctx context.Context, familyID string, days int) ([]*DailyGameTime, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DATE(started_at)::TEXT AS day, COALESCE(SUM(duration_seconds) / 60, 0) AS minutes
		 FROM game_sessions
		 WHERE family_id = $1 AND started_at >= NOW() - ($2 || ' days')::INTERVAL
		 GROUP BY day ORDER BY day ASC`,
		familyID, days,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*DailyGameTime
	for rows.Next() {
		var d DailyGameTime
		if err := rows.Scan(&d.Date, &d.Minutes); err != nil {
			return nil, err
		}
		result = append(result, &d)
	}
	return result, rows.Err()
}

type QuizScoreEntry struct {
	Date     string  `json:"date"`
	QuizType string  `json:"quiz_type"`
	AvgScore float64 `json:"avg_score"`
}

func (r *DashboardRepo) QuizScores(ctx context.Context, familyID string, days int) ([]*QuizScoreEntry, error) {
	cutoff := time.Now().AddDate(0, 0, -days)
	var result []*QuizScoreEntry

	// Hadith quizzes
	rows, err := r.db.QueryContext(ctx,
		`SELECT DATE(completed_at)::TEXT AS day, 'hadith' AS quiz_type, AVG(score) AS avg_score
		 FROM hadith_quizzes
		 WHERE family_id = $1 AND status = 'completed' AND completed_at >= $2
		 GROUP BY day ORDER BY day ASC`,
		familyID, cutoff,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var e QuizScoreEntry
		if err := rows.Scan(&e.Date, &e.QuizType, &e.AvgScore); err != nil {
			return nil, err
		}
		result = append(result, &e)
	}

	// Prophet quizzes
	rows2, err := r.db.QueryContext(ctx,
		`SELECT DATE(completed_at)::TEXT AS day, 'prophet' AS quiz_type, AVG(score) AS avg_score
		 FROM prophet_quizzes
		 WHERE family_id = $1 AND status = 'completed' AND completed_at >= $2
		 GROUP BY day ORDER BY day ASC`,
		familyID, cutoff,
	)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var e QuizScoreEntry
		if err := rows2.Scan(&e.Date, &e.QuizType, &e.AvgScore); err != nil {
			return nil, err
		}
		result = append(result, &e)
	}

	// Quran quizzes
	rows3, err := r.db.QueryContext(ctx,
		`SELECT DATE(completed_at)::TEXT AS day, 'quran' AS quiz_type, AVG(score) AS avg_score
		 FROM quran_quizzes
		 WHERE family_id = $1 AND status = 'completed' AND completed_at >= $2
		 GROUP BY day ORDER BY day ASC`,
		familyID, cutoff,
	)
	if err != nil {
		return nil, err
	}
	defer rows3.Close()
	for rows3.Next() {
		var e QuizScoreEntry
		if err := rows3.Scan(&e.Date, &e.QuizType, &e.AvgScore); err != nil {
			return nil, err
		}
		result = append(result, &e)
	}

	return result, nil
}

type LearnProgressEntry struct {
	ContentID     string  `json:"content_id"`
	Title         string  `json:"title"`
	ContentType   string  `json:"content_type"`
	CompletedBy   int     `json:"completed_by"`
	TotalAssigned int     `json:"total_assigned"`
	ProgressPct   float64 `json:"progress_pct"`
}

func (r *DashboardRepo) LearnProgress(ctx context.Context, familyID string) ([]*LearnProgressEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
		   lc.id::TEXT,
		   lc.title,
		   lc.content_type,
		   COUNT(lp.id) FILTER (WHERE lp.completed_at IS NOT NULL) AS completed_by,
		   COUNT(u.id) AS total_assigned
		 FROM learn_content lc
		 LEFT JOIN users u ON u.family_id = lc.family_id AND (lc.assigned_to IS NULL OR lc.assigned_to = u.id)
		 LEFT JOIN learn_progress lp ON lp.content_id = lc.id AND lp.user_id = u.id
		 WHERE lc.family_id = $1
		 GROUP BY lc.id, lc.title, lc.content_type
		 ORDER BY lc.created_at DESC`,
		familyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*LearnProgressEntry
	for rows.Next() {
		var e LearnProgressEntry
		if err := rows.Scan(&e.ContentID, &e.Title, &e.ContentType, &e.CompletedBy, &e.TotalAssigned); err != nil {
			return nil, err
		}
		if e.TotalAssigned > 0 {
			e.ProgressPct = float64(e.CompletedBy) / float64(e.TotalAssigned) * 100
		}
		result = append(result, &e)
	}
	return result, rows.Err()
}
