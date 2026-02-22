package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type ProphetRepo struct {
	db *sqlx.DB
}

func NewProphetRepo(db *sqlx.DB) *ProphetRepo {
	return &ProphetRepo{db: db}
}

func (r *ProphetRepo) List(ctx context.Context) ([]*models.Prophet, error) {
	var prophets []*models.Prophet
	err := r.db.SelectContext(ctx, &prophets,
		`SELECT id, name_en, name_ar, order_num, story_summary, key_miracles, nation, quran_refs, difficulty, created_at
		 FROM prophets ORDER BY order_num ASC NULLS LAST, name_en ASC`,
	)
	return prophets, err
}

func (r *ProphetRepo) GetByID(ctx context.Context, id string) (*models.Prophet, error) {
	var p models.Prophet
	err := r.db.GetContext(ctx, &p,
		`SELECT id, name_en, name_ar, order_num, story_summary, key_miracles, nation, quran_refs, difficulty, created_at
		 FROM prophets WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProphetRepo) GetByOrderNum(ctx context.Context, n int) (*models.Prophet, error) {
	var p models.Prophet
	err := r.db.GetContext(ctx, &p,
		`SELECT id, name_en, name_ar, order_num, story_summary, key_miracles, nation, quran_refs, difficulty, created_at
		 FROM prophets WHERE order_num = $1`,
		n,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
