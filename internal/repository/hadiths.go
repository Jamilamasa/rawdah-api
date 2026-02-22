package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type HadithRepo struct {
	db *sqlx.DB
}

func NewHadithRepo(db *sqlx.DB) *HadithRepo {
	return &HadithRepo{db: db}
}

func (r *HadithRepo) List(ctx context.Context, difficulty string) ([]*models.Hadith, error) {
	query := `SELECT id, text_en, text_ar, source, topic, difficulty, created_at FROM hadiths`
	args := []interface{}{}

	if difficulty != "" {
		query += " WHERE difficulty = $1"
		args = append(args, difficulty)
	}
	query += " ORDER BY created_at DESC"

	var hadiths []*models.Hadith
	err := r.db.SelectContext(ctx, &hadiths, query, args...)
	return hadiths, err
}

func (r *HadithRepo) GetByID(ctx context.Context, id string) (*models.Hadith, error) {
	var h models.Hadith
	err := r.db.GetContext(ctx, &h,
		`SELECT id, text_en, text_ar, source, topic, difficulty, created_at FROM hadiths WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *HadithRepo) Random(ctx context.Context, difficulty string) (*models.Hadith, error) {
	query := `SELECT id, text_en, text_ar, source, topic, difficulty, created_at FROM hadiths`
	args := []interface{}{}

	if difficulty != "" {
		query += " WHERE difficulty = $1"
		args = append(args, difficulty)
		query += " ORDER BY RANDOM() LIMIT 1"
	} else {
		query += " ORDER BY RANDOM() LIMIT 1"
	}

	var h models.Hadith
	err := r.db.GetContext(ctx, &h, query, args...)
	if err != nil {
		return nil, fmt.Errorf("no hadith found")
	}
	return &h, nil
}
