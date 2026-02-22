package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type GameRepo struct {
	db *sqlx.DB
}

func NewGameRepo(db *sqlx.DB) *GameRepo {
	return &GameRepo{db: db}
}

func (r *GameRepo) StartSession(ctx context.Context, session *models.GameSession) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO game_sessions (user_id, family_id, game_name, game_type, started_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		session.UserID, session.FamilyID, session.GameName, session.GameType, session.StartedAt,
	).Scan(&session.ID)
}

func (r *GameRepo) EndSession(ctx context.Context, id, userID, familyID string, durationSeconds int) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE game_sessions SET ended_at = $1, duration_seconds = $2
		 WHERE id = $3 AND user_id = $4 AND family_id = $5 AND ended_at IS NULL`,
		now, durationSeconds, id, userID, familyID,
	)
	return err
}

func (r *GameRepo) TotalDurationToday(ctx context.Context, userID, familyID string, date time.Time) (int, error) {
	var total *int
	err := r.db.QueryRowContext(ctx,
		`SELECT SUM(duration_seconds) FROM game_sessions
		 WHERE user_id = $1 AND family_id = $2 AND DATE(started_at) = $3`,
		userID, familyID, date.Format("2006-01-02"),
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	if total == nil {
		return 0, nil
	}
	return *total, nil
}

func (r *GameRepo) ListSessions(ctx context.Context, familyID, userID string) ([]*models.GameSession, error) {
	query := `SELECT id, user_id, family_id, game_name, game_type, started_at, ended_at, duration_seconds
	          FROM game_sessions WHERE family_id = $1`
	args := []interface{}{familyID}

	if userID != "" {
		query += " AND user_id = $2"
		args = append(args, userID)
	}
	query += " ORDER BY started_at DESC"

	var sessions []*models.GameSession
	err := r.db.SelectContext(ctx, &sessions, query, args...)
	return sessions, err
}

func (r *GameRepo) CountSessions(ctx context.Context, userID, familyID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM game_sessions WHERE user_id = $1 AND family_id = $2`,
		userID, familyID,
	).Scan(&count)
	return count, err
}
