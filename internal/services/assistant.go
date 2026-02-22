package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rawdah/rawdah-api/internal/ai"
	"github.com/rawdah/rawdah-api/internal/repository"
)

const (
	defaultAssistantChildAge = 10
	minAssistantQuestionLen  = 2
	maxAssistantQuestionLen  = 2000
)

var ErrInvalidAssistantQuestion = errors.New("invalid assistant question")

type AssistantService struct {
	aiClient   *ai.Client
	familyRepo *repository.FamilyRepo
}

type AskAssistantInput struct {
	FamilyID string
	UserID   string
	Role     string
	Question string
}

func NewAssistantService(aiClient *ai.Client, familyRepo *repository.FamilyRepo) *AssistantService {
	return &AssistantService{
		aiClient:   aiClient,
		familyRepo: familyRepo,
	}
}

func (s *AssistantService) Ask(ctx context.Context, input AskAssistantInput) (string, error) {
	question := strings.TrimSpace(input.Question)
	if len(question) < minAssistantQuestionLen || len(question) > maxAssistantQuestionLen {
		return "", ErrInvalidAssistantQuestion
	}

	systemPrompt := ai.BuildParentAssistantPrompt()
	if input.Role == "child" {
		member, err := s.familyRepo.GetMemberByID(ctx, input.UserID, input.FamilyID)
		if err != nil {
			return "", fmt.Errorf("failed to load child profile: %w", err)
		}
		systemPrompt = ai.BuildKidAssistantPrompt(childAgeFromDOB(member.DateOfBirth))
	}

	answer, err := s.aiClient.AskQuestion(systemPrompt, question)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(answer), nil
}

func childAgeFromDOB(dateOfBirth *time.Time) int {
	if dateOfBirth == nil {
		return defaultAssistantChildAge
	}

	now := time.Now().UTC()
	age := now.Year() - dateOfBirth.Year()
	if now.Month() < dateOfBirth.Month() || (now.Month() == dateOfBirth.Month() && now.Day() < dateOfBirth.Day()) {
		age--
	}

	if age < minChildAge || age > maxChildAge {
		return defaultAssistantChildAge
	}
	return age
}
