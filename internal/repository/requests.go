package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/models"
)

type RequestRepo struct {
	db *sqlx.DB
}

func NewRequestRepo(db *sqlx.DB) *RequestRepo {
	return &RequestRepo{db: db}
}

func (r *RequestRepo) Create(ctx context.Context, req *models.Request) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO requests (family_id, requester_id, target_id, title, description, status)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		req.FamilyID, req.RequesterID, req.TargetID, req.Title, req.Description, req.Status,
	).Scan(&req.ID, &req.CreatedAt)
}

func (r *RequestRepo) List(ctx context.Context, familyID string) ([]*models.Request, error) {
	var requests []*models.Request
	err := r.db.SelectContext(ctx, &requests,
		`SELECT id, family_id, requester_id, target_id, title, description, status,
		        response_message, responded_by, responded_at, created_at
		 FROM requests WHERE family_id = $1 ORDER BY created_at DESC`,
		familyID,
	)
	return requests, err
}

func (r *RequestRepo) GetByID(ctx context.Context, id, familyID string) (*models.Request, error) {
	var req models.Request
	err := r.db.GetContext(ctx, &req,
		`SELECT id, family_id, requester_id, target_id, title, description, status,
		        response_message, responded_by, responded_at, created_at
		 FROM requests WHERE id = $1 AND family_id = $2`,
		id, familyID,
	)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *RequestRepo) Respond(ctx context.Context, id, familyID, status string, message *string, respondedBy uuid.UUID) (*models.Request, error) {
	now := time.Now()
	var req models.Request
	err := r.db.GetContext(ctx, &req,
		`UPDATE requests SET status = $1, response_message = $2, responded_by = $3, responded_at = $4
		 WHERE id = $5 AND family_id = $6
		 RETURNING id, family_id, requester_id, target_id, title, description, status,
		           response_message, responded_by, responded_at, created_at`,
		status, message, respondedBy, now, id, familyID,
	)
	if err != nil {
		return nil, fmt.Errorf("request not found")
	}
	return &req, nil
}
