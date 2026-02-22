package mailer

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/rawdah/rawdah-api/internal/config"
)

type BrevoContact struct {
	Name  string
	Email string
}

type Mailer struct {
	cfg    *config.Config
	client *resty.Client
}

func NewMailer(cfg *config.Config) *Mailer {
	client := resty.New().
		SetBaseURL("https://api.brevo.com/v3").
		SetHeader("api-key", cfg.BrevoAPIKey).
		SetHeader("Content-Type", "application/json")

	return &Mailer{cfg: cfg, client: client}
}

func (m *Mailer) Send(to BrevoContact, subject, htmlBody string) error {
	payload := map[string]interface{}{
		"sender": map[string]string{
			"name":  m.cfg.BrevoSenderName,
			"email": m.cfg.BrevoSenderEmail,
		},
		"to": []map[string]string{
			{"name": to.Name, "email": to.Email},
		},
		"subject":     subject,
		"htmlContent": htmlBody,
	}

	resp, err := m.client.R().SetBody(payload).Post("/smtp/email")
	if err != nil {
		return fmt.Errorf("brevo send error: %w", err)
	}
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("brevo API error %d: %s", resp.StatusCode(), resp.String())
	}
	return nil
}
