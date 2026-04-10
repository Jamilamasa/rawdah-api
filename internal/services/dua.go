package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-resty/resty/v2"
	"github.com/rawdah/rawdah-api/internal/models"
	"github.com/rawdah/rawdah-api/internal/repository"
)

const (
	maxDuaNarrativeFieldLen = 2000
	maxDuaFormatFieldLen    = 120
	defaultDuaProviderURL   = "https://dua-companion-api.onrender.com"
	defaultDuaTimeout       = 20 * time.Second
	maxDuaHistoryLimit      = 100
	defaultDuaHistoryLimit  = 20
)

var (
	ErrInvalidDuaInput        = errors.New("invalid dua input")
	ErrDuaProviderUnavailable = errors.New("dua provider unavailable")
	ErrDuaHistoryNotFound     = errors.New("dua history entry not found")
)

type DuaService struct {
	client      *resty.Client
	historyRepo *repository.DuaHistoryRepo
}

type GenerateDuaInput struct {
	FamilyID     string
	UserID       string
	AskingFor    string
	HeavyOnHeart string
	AfraidOf     string
	IfAnswered   string
	OutputStyle  string
	Depth        string
	Tone         string
}

type DuaSelectedName = models.DuaSelectedName

type DuaGenerateResponse struct {
	SelectedNames []DuaSelectedName `json:"selected_names"`
	DuaText       string            `json:"dua_text"`
}

type ListDuaHistoryInput struct {
	FamilyID string
	UserID   string
	Limit    int
}

func NewDuaService(baseURL string, timeout time.Duration, historyRepo *repository.DuaHistoryRepo) *DuaService {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = defaultDuaProviderURL
	}
	if timeout <= 0 {
		timeout = defaultDuaTimeout
	}

	client := resty.New().
		SetBaseURL(strings.TrimRight(baseURL, "/")).
		SetTimeout(timeout).
		SetHeader("Content-Type", "application/json")

	return &DuaService{
		client:      client,
		historyRepo: historyRepo,
	}
}

func (s *DuaService) Generate(ctx context.Context, input GenerateDuaInput) (*DuaGenerateResponse, error) {
	familyID, err := parseUUID(strings.TrimSpace(input.FamilyID))
	if err != nil {
		return nil, fmt.Errorf("%w: family_id is invalid", ErrInvalidDuaInput)
	}
	userID, err := parseUUID(strings.TrimSpace(input.UserID))
	if err != nil {
		return nil, fmt.Errorf("%w: user_id is invalid", ErrInvalidDuaInput)
	}

	if s.historyRepo == nil {
		return nil, fmt.Errorf("%w: history repository is not configured", ErrDuaProviderUnavailable)
	}

	input.AskingFor, err = sanitizeDuaField(input.AskingFor, "asking_for", maxDuaNarrativeFieldLen)
	if err != nil {
		return nil, err
	}
	input.HeavyOnHeart, err = sanitizeDuaField(input.HeavyOnHeart, "heavy_on_heart", maxDuaNarrativeFieldLen)
	if err != nil {
		return nil, err
	}
	input.AfraidOf, err = sanitizeDuaField(input.AfraidOf, "afraid_of", maxDuaNarrativeFieldLen)
	if err != nil {
		return nil, err
	}
	input.IfAnswered, err = sanitizeDuaField(input.IfAnswered, "if_answered", maxDuaNarrativeFieldLen)
	if err != nil {
		return nil, err
	}
	input.OutputStyle, err = sanitizeDuaField(input.OutputStyle, "output_style", maxDuaFormatFieldLen)
	if err != nil {
		return nil, err
	}
	input.Depth, err = sanitizeDuaField(input.Depth, "depth", maxDuaFormatFieldLen)
	if err != nil {
		return nil, err
	}
	input.Tone, err = sanitizeDuaField(input.Tone, "tone", maxDuaFormatFieldLen)
	if err != nil {
		return nil, err
	}

	payload := map[string]string{
		"asking_for":     input.AskingFor,
		"heavy_on_heart": input.HeavyOnHeart,
		"afraid_of":      input.AfraidOf,
		"if_answered":    input.IfAnswered,
		"output_style":   input.OutputStyle,
		"depth":          input.Depth,
		"tone":           input.Tone,
	}

	var out DuaGenerateResponse
	httpResp, err := s.client.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&out).
		Post("/generate-dua")
	if err != nil {
		return nil, fmt.Errorf("%w: request failed: %v", ErrDuaProviderUnavailable, err)
	}

	if httpResp.StatusCode() >= 500 {
		return nil, fmt.Errorf("%w: upstream returned %d", ErrDuaProviderUnavailable, httpResp.StatusCode())
	}
	if httpResp.StatusCode() >= 400 {
		return nil, fmt.Errorf("%w: provider rejected request with status %d", ErrInvalidDuaInput, httpResp.StatusCode())
	}
	if strings.TrimSpace(out.DuaText) == "" {
		return nil, fmt.Errorf("%w: empty dua_text in upstream response", ErrDuaProviderUnavailable)
	}

	historyEntry := &models.DuaHistory{
		FamilyID:      familyID,
		UserID:        userID,
		AskingFor:     input.AskingFor,
		HeavyOnHeart:  input.HeavyOnHeart,
		AfraidOf:      input.AfraidOf,
		IfAnswered:    input.IfAnswered,
		OutputStyle:   input.OutputStyle,
		Depth:         input.Depth,
		Tone:          input.Tone,
		SelectedNames: out.SelectedNames,
		DuaText:       strings.TrimSpace(out.DuaText),
	}
	if err := s.historyRepo.Create(ctx, historyEntry); err != nil {
		return nil, fmt.Errorf("save dua history: %w", err)
	}

	return &out, nil
}

func (s *DuaService) ListHistory(ctx context.Context, input ListDuaHistoryInput) ([]*models.DuaHistory, error) {
	if s.historyRepo == nil {
		return nil, fmt.Errorf("%w: history repository is not configured", ErrDuaProviderUnavailable)
	}
	if _, err := parseUUID(strings.TrimSpace(input.FamilyID)); err != nil {
		return nil, fmt.Errorf("%w: family_id is invalid", ErrInvalidDuaInput)
	}
	if _, err := parseUUID(strings.TrimSpace(input.UserID)); err != nil {
		return nil, fmt.Errorf("%w: user_id is invalid", ErrInvalidDuaInput)
	}

	limit := input.Limit
	switch {
	case limit <= 0:
		limit = defaultDuaHistoryLimit
	case limit > maxDuaHistoryLimit:
		return nil, fmt.Errorf("%w: limit must be between 1 and %d", ErrInvalidDuaInput, maxDuaHistoryLimit)
	}

	history, err := s.historyRepo.ListByUser(ctx, input.FamilyID, input.UserID, limit)
	if err != nil {
		return nil, fmt.Errorf("list dua history: %w", err)
	}
	return history, nil
}

func (s *DuaService) GetHistory(ctx context.Context, id, familyID, userID string) (*models.DuaHistory, error) {
	if s.historyRepo == nil {
		return nil, fmt.Errorf("%w: history repository is not configured", ErrDuaProviderUnavailable)
	}
	if _, err := parseUUID(strings.TrimSpace(id)); err != nil {
		return nil, fmt.Errorf("%w: id is invalid", ErrInvalidDuaInput)
	}
	if _, err := parseUUID(strings.TrimSpace(familyID)); err != nil {
		return nil, fmt.Errorf("%w: family_id is invalid", ErrInvalidDuaInput)
	}
	if _, err := parseUUID(strings.TrimSpace(userID)); err != nil {
		return nil, fmt.Errorf("%w: user_id is invalid", ErrInvalidDuaInput)
	}

	entry, err := s.historyRepo.GetByIDForUser(ctx, id, familyID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDuaHistoryNotFound
		}
		return nil, fmt.Errorf("get dua history: %w", err)
	}
	return entry, nil
}

func sanitizeDuaField(value, field string, maxLen int) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%w: %s is required", ErrInvalidDuaInput, field)
	}
	if utf8.RuneCountInString(value) > maxLen {
		return "", fmt.Errorf("%w: %s must be at most %d characters", ErrInvalidDuaInput, field, maxLen)
	}
	return value, nil
}
