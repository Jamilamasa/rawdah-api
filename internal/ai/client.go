package ai

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/rawdah/rawdah-api/internal/config"
	"github.com/rawdah/rawdah-api/internal/models"
)

type Client struct {
	cfg   *config.Config
	resty *resty.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg:   cfg,
		resty: resty.New().SetBaseURL("https://openrouter.ai/api/v1"),
	}
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterRequest struct {
	Model    string              `json:"model"`
	Messages []openRouterMessage `json:"messages"`
}

type openRouterChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type openRouterResponse struct {
	Choices []openRouterChoice `json:"choices"`
}

func (c *Client) GenerateQuiz(prompt string) ([]models.QuizQuestion, error) {
	questions, err := c.callModel(c.cfg.OpenRouterModel, prompt)
	if err != nil {
		// Fallback to second model
		questions, err = c.callModel(c.cfg.OpenRouterFallbackModel, prompt)
		if err != nil {
			return nil, fmt.Errorf("both AI models failed: %w", err)
		}
	}
	return questions, nil
}

func (c *Client) callModel(model, prompt string) ([]models.QuizQuestion, error) {
	req := openRouterRequest{
		Model: model,
		Messages: []openRouterMessage{
			{Role: "user", Content: prompt},
		},
	}

	var resp openRouterResponse
	httpResp, err := c.resty.R().
		SetHeader("Authorization", "Bearer "+c.cfg.OpenRouterAPIKey).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&resp).
		Post("/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	if httpResp.StatusCode() >= 400 {
		return nil, fmt.Errorf("API error %d: %s", httpResp.StatusCode(), httpResp.String())
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := resp.Choices[0].Message.Content

	// Extract JSON from response (model may wrap in markdown code blocks)
	jsonStr := extractJSON(content)

	var questions []models.QuizQuestion
	if err := json.Unmarshal([]byte(jsonStr), &questions); err != nil {
		return nil, fmt.Errorf("failed to parse quiz JSON: %w, content: %s", err, content)
	}

	if err := validateQuestions(questions); err != nil {
		return nil, err
	}

	return questions, nil
}

func extractJSON(content string) string {
	// Try to find JSON array in the content
	start := -1
	end := -1
	for i, ch := range content {
		if ch == '[' && start == -1 {
			start = i
		}
		if ch == ']' {
			end = i
		}
	}
	if start != -1 && end != -1 && end > start {
		return content[start : end+1]
	}
	return content
}

func validateQuestions(questions []models.QuizQuestion) error {
	if len(questions) < 3 {
		return fmt.Errorf("expected at least 3 questions, got %d", len(questions))
	}
	for i, q := range questions {
		if q.ID == "" {
			return fmt.Errorf("question %d missing id", i)
		}
		if q.Question == "" {
			return fmt.Errorf("question %d missing question text", i)
		}
		if len(q.Options) != 4 {
			return fmt.Errorf("question %d must have exactly 4 options", i)
		}
		for _, key := range []string{"A", "B", "C", "D"} {
			if _, ok := q.Options[key]; !ok {
				return fmt.Errorf("question %d missing option %s", i, key)
			}
		}
		if q.CorrectAnswer == "" {
			return fmt.Errorf("question %d missing correct_answer", i)
		}
		if q.Explanation == "" {
			return fmt.Errorf("question %d missing explanation", i)
		}
	}
	return nil
}
