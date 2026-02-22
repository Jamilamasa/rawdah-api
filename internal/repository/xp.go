package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type XPRepo struct {
	db *sqlx.DB
}

func NewXPRepo(db *sqlx.DB) *XPRepo {
	return &XPRepo{db: db}
}

func (r *XPRepo) GetOrCreateUserXP(ctx context.Context, userID, familyID string) (*models.UserXP, error) {
	var xp models.UserXP
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO user_xp (user_id, family_id, total_xp, level)
		 VALUES ($1, $2, 0, 1)
		 ON CONFLICT (user_id) DO UPDATE SET user_id = EXCLUDED.user_id
		 RETURNING id, user_id, family_id, total_xp, level, updated_at`,
		userID, familyID,
	).Scan(&xp.ID, &xp.UserID, &xp.FamilyID, &xp.TotalXP, &xp.Level, &xp.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &xp, nil
}

func (r *XPRepo) AddXP(ctx context.Context, userID, familyID string, amount int) (*models.UserXP, error) {
	var xp models.UserXP
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO user_xp (user_id, family_id, total_xp, level, updated_at)
		 VALUES ($1, $2, $3, 1, NOW())
		 ON CONFLICT (user_id) DO UPDATE
		   SET total_xp = user_xp.total_xp + $3, updated_at = NOW()
		 RETURNING id, user_id, family_id, total_xp, level, updated_at`,
		userID, familyID, amount,
	).Scan(&xp.ID, &xp.UserID, &xp.FamilyID, &xp.TotalXP, &xp.Level, &xp.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Update level
	newLevel := r.GetLevel(xp.TotalXP)
	if newLevel != xp.Level {
		_, err = r.db.ExecContext(ctx,
			`UPDATE user_xp SET level = $1, updated_at = NOW() WHERE user_id = $2`,
			newLevel, userID,
		)
		if err != nil {
			return nil, err
		}
		xp.Level = newLevel
	}
	return &xp, nil
}

func (r *XPRepo) GetLevel(totalXP int) int {
	level := 1
	for _, t := range models.LevelThresholds {
		if totalXP >= t.XP {
			level = t.Level
		}
	}
	return level
}

func (r *XPRepo) InsertXPEvent(ctx context.Context, userID, familyID, source string, sourceID uuid.UUID, amount int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO xp_events (user_id, family_id, source, source_id, xp_amount)
		 VALUES ($1, $2, $3, $4, $5)`,
		userID, familyID, source, sourceID, amount,
	)
	return err
}

func (r *XPRepo) CountCompletedTasks(ctx context.Context, userID, familyID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tasks
		 WHERE assigned_to = $1
		   AND family_id = $2
		   AND status IN ('completed', 'reward_requested', 'reward_approved')`,
		userID, familyID,
	).Scan(&count)
	return count, err
}

func (r *XPRepo) CountGameSessions(ctx context.Context, userID, familyID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM game_sessions WHERE user_id = $1 AND family_id = $2`,
		userID, familyID,
	).Scan(&count)
	return count, err
}

func (r *XPRepo) CountMessagesSent(ctx context.Context, userID, familyID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM messages WHERE sender_id = $1 AND family_id = $2`,
		userID, familyID,
	).Scan(&count)
	return count, err
}

func (r *XPRepo) CheckStreak(ctx context.Context, userID, familyID string, days int) (bool, error) {
	// Check if user has XP events on each of the last `days` days
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT DATE(created_at)) FROM xp_events
		 WHERE user_id = $1
		   AND family_id = $2
		   AND created_at >= NOW() - ($3 || ' days')::INTERVAL`,
		userID, familyID, days,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count >= days, nil
}

func (r *XPRepo) GetBadge(ctx context.Context, slug string) (*models.Badge, error) {
	var b models.Badge
	err := r.db.GetContext(ctx, &b,
		`SELECT id, slug, name, description, icon, xp_reward FROM badges WHERE slug = $1`,
		slug,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *XPRepo) HasBadge(ctx context.Context, userID string, badgeID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_badges WHERE user_id = $1 AND badge_id = $2)`,
		userID, badgeID,
	).Scan(&exists)
	return exists, err
}

func (r *XPRepo) AwardBadge(ctx context.Context, userID string, badgeID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_badges (user_id, badge_id, awarded_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, badge_id) DO NOTHING`,
		userID, badgeID, time.Now(),
	)
	return err
}

func (r *XPRepo) GetUserXP(ctx context.Context, userID, familyID string) (*models.UserXP, error) {
	var xp models.UserXP
	err := r.db.GetContext(ctx, &xp,
		`SELECT id, user_id, family_id, total_xp, level, updated_at
		 FROM user_xp WHERE user_id = $1 AND family_id = $2`,
		userID, familyID,
	)
	if err != nil {
		return nil, err
	}
	return &xp, nil
}
