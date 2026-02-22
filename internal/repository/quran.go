package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type QuranRepo struct {
	db *sqlx.DB
}

func NewQuranRepo(db *sqlx.DB) *QuranRepo {
	return &QuranRepo{db: db}
}

func (r *QuranRepo) ListVerses(ctx context.Context, topic, difficulty string) ([]*models.QuranVerse, error) {
	query := `SELECT id, surah_number, ayah_number, surah_name_en, text_ar, text_en,
	                 transliteration, tafsir_simple, life_application, topic, difficulty, created_at
	          FROM quran_verses WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if topic != "" {
		query += fmt.Sprintf(" AND topic = $%d", argIdx)
		args = append(args, topic)
		argIdx++
	}
	if difficulty != "" {
		query += fmt.Sprintf(" AND difficulty = $%d", argIdx)
		args = append(args, difficulty)
		argIdx++
	}
	query += " ORDER BY surah_number ASC, ayah_number ASC"

	var verses []*models.QuranVerse
	err := r.db.SelectContext(ctx, &verses, query, args...)
	return verses, err
}

func (r *QuranRepo) GetVerseByID(ctx context.Context, id string) (*models.QuranVerse, error) {
	var v models.QuranVerse
	err := r.db.GetContext(ctx, &v,
		`SELECT id, surah_number, ayah_number, surah_name_en, text_ar, text_en,
		        transliteration, tafsir_simple, life_application, topic, difficulty, created_at
		 FROM quran_verses WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
