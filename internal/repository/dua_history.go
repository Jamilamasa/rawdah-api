package repository

import (
	"context"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

const defaultDuaHistoryLimit = 20

type DuaHistoryRepo struct {
	db *sqlx.DB
}

func NewDuaHistoryRepo(db *sqlx.DB) *DuaHistoryRepo {
	return &DuaHistoryRepo{db: db}
}

func (r *DuaHistoryRepo) Create(ctx context.Context, history *models.DuaHistory) error {
	selectedNamesJSON, err := json.Marshal(history.SelectedNames)
	if err != nil {
		return err
	}

	return r.db.QueryRowContext(ctx,
		`INSERT INTO dua_history (
			family_id,
			user_id,
			asking_for,
			heavy_on_heart,
			afraid_of,
			if_answered,
			output_style,
			depth,
			tone,
			selected_names,
			dua_text
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
		RETURNING id, created_at`,
		history.FamilyID,
		history.UserID,
		history.AskingFor,
		history.HeavyOnHeart,
		history.AfraidOf,
		history.IfAnswered,
		history.OutputStyle,
		history.Depth,
		history.Tone,
		selectedNamesJSON,
		history.DuaText,
	).Scan(&history.ID, &history.CreatedAt)
}

func (r *DuaHistoryRepo) ListByUser(ctx context.Context, familyID, userID string, limit int) ([]*models.DuaHistory, error) {
	if limit <= 0 {
		limit = defaultDuaHistoryLimit
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT
			id,
			family_id,
			user_id,
			asking_for,
			heavy_on_heart,
			afraid_of,
			if_answered,
			output_style,
			depth,
			tone,
			selected_names,
			dua_text,
			created_at
		FROM dua_history
		WHERE family_id = $1 AND user_id = $2
		ORDER BY created_at DESC
		LIMIT $3`,
		familyID,
		userID,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]*models.DuaHistory, 0)
	for rows.Next() {
		item, err := scanDuaHistory(rows)
		if err != nil {
			return nil, err
		}
		history = append(history, item)
	}
	return history, rows.Err()
}

func (r *DuaHistoryRepo) GetByIDForUser(ctx context.Context, id, familyID, userID string) (*models.DuaHistory, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT
			id,
			family_id,
			user_id,
			asking_for,
			heavy_on_heart,
			afraid_of,
			if_answered,
			output_style,
			depth,
			tone,
			selected_names,
			dua_text,
			created_at
		FROM dua_history
		WHERE id = $1 AND family_id = $2 AND user_id = $3`,
		id,
		familyID,
		userID,
	)
	return scanDuaHistory(row)
}

func scanDuaHistory(row interface{ Scan(...interface{}) error }) (*models.DuaHistory, error) {
	var item models.DuaHistory
	var selectedNamesRaw []byte
	err := row.Scan(
		&item.ID,
		&item.FamilyID,
		&item.UserID,
		&item.AskingFor,
		&item.HeavyOnHeart,
		&item.AfraidOf,
		&item.IfAnswered,
		&item.OutputStyle,
		&item.Depth,
		&item.Tone,
		&selectedNamesRaw,
		&item.DuaText,
		&item.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(selectedNamesRaw) > 0 {
		if err := json.Unmarshal(selectedNamesRaw, &item.SelectedNames); err != nil {
			return nil, err
		}
	} else {
		item.SelectedNames = []models.DuaSelectedName{}
	}

	return &item, nil
}
