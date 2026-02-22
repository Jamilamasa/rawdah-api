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

type openRouterResponseFormat struct {
	Type string `json:"type"`
}

type openRouterRequest struct {
	Model          string                   `json:"model"`
	Messages       []openRouterMessage      `json:"messages"`
	ResponseFormat openRouterResponseFormat `json:"response_format"`
}

type openRouterChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type openRouterResponse struct {
	Choices []openRouterChoice `json:"choices"`
}

// AIHadith holds the hadith content returned by the AI when generating a hadith quiz.
type AIHadith struct {
	TextEn string `json:"text_en"`
	TextAr string `json:"text_ar"`
	Source string `json:"source"`
	Topic  string `json:"topic"`
}

// HadithQuizResult is the full response from the AI for a hadith quiz.
type HadithQuizResult struct {
	Hadith    AIHadith
	Questions []models.QuizQuestion
}

// GenerateHadithQuiz calls the AI with a self-contained prompt that selects an authentic
// hadith and generates quiz questions about it. Returns both the hadith and the questions.
func (c *Client) GenerateHadithQuiz(prompt string) (*HadithQuizResult, error) {
	result, err := c.callHadithModel(c.cfg.OpenRouterModel, prompt)
	if err != nil {
		result, err = c.callHadithModel(c.cfg.OpenRouterFallbackModel, prompt)
		if err != nil {
			return nil, fmt.Errorf("both AI models failed: %w", err)
		}
	}
	return result, nil
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

func (c *Client) callHadithModel(model, prompt string) (*HadithQuizResult, error) {
	req := openRouterRequest{
		Model:          model,
		Messages:       []openRouterMessage{{Role: "user", Content: prompt}},
		ResponseFormat: openRouterResponseFormat{Type: "json_object"},
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

	jsonStr := extractJSONObject(resp.Choices[0].Message.Content)

	var wrapper struct {
		Hadith    AIHadith              `json:"hadith"`
		Questions []models.QuizQuestion `json:"questions"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse hadith quiz JSON: %w, content: %s", err, resp.Choices[0].Message.Content)
	}

	if wrapper.Hadith.TextEn == "" {
		return nil, fmt.Errorf("AI returned empty hadith text")
	}
	if wrapper.Hadith.Source == "" {
		return nil, fmt.Errorf("AI returned hadith with no source")
	}
	if err := validateQuestions(wrapper.Questions); err != nil {
		return nil, err
	}

	return &HadithQuizResult{Hadith: wrapper.Hadith, Questions: wrapper.Questions}, nil
}

func (c *Client) callModel(model, prompt string) ([]models.QuizQuestion, error) {
	req := openRouterRequest{
		Model: model,
		Messages: []openRouterMessage{
			{Role: "user", Content: prompt},
		},
		ResponseFormat: openRouterResponseFormat{Type: "json_object"},
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

	// Extract the JSON object from the response (model may wrap in markdown code blocks)
	jsonStr := extractJSONObject(content)

	var wrapper struct {
		Questions []models.QuizQuestion `json:"questions"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse quiz JSON: %w, content: %s", err, content)
	}

	if err := validateQuestions(wrapper.Questions); err != nil {
		return nil, err
	}

	return wrapper.Questions, nil
}

func extractJSONObject(content string) string {
	start := -1
	depth := 0
	for i, ch := range content {
		if ch == '{' {
			if start == -1 {
				start = i
			}
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 && start != -1 {
				return content[start : i+1]
			}
		}
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
