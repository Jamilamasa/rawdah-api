package push

import (
	"context"
	"encoding/json"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/rawdah/rawdah-api/internal/config"
	"github.com/rawdah/rawdah-api/internal/models"
)

type PushPayload struct {
	Title string      `json:"title"`
	Body  string      `json:"body"`
	Data  interface{} `json:"data,omitempty"`
	Icon  string      `json:"icon,omitempty"`
	URL   string      `json:"url,omitempty"`
}

type SubGetter interface {
	GetSubscriptions(ctx context.Context, userID string) ([]*models.PushSubscription, error)
}

type PushSender struct {
	cfg  *config.Config
	repo SubGetter
}

func NewPushSender(cfg *config.Config, repo SubGetter) *PushSender {
	return &PushSender{cfg: cfg, repo: repo}
}

func (p *PushSender) SendToUser(ctx context.Context, userID string, payload PushPayload) {
	subs, err := p.repo.GetSubscriptions(ctx, userID)
	if err != nil || len(subs) == 0 {
		return
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	for _, sub := range subs {
		sub := sub // capture
		go func() {
			s := &webpush.Subscription{
				Endpoint: sub.Endpoint,
				Keys: webpush.Keys{
					P256dh: sub.P256dh,
					Auth:   sub.Auth,
				},
			}
			_, _ = webpush.SendNotification(payloadBytes, s, &webpush.Options{
				VAPIDPublicKey:  p.cfg.VAPIDPublicKey,
				VAPIDPrivateKey: p.cfg.VAPIDPrivateKey,
				Subscriber:      p.cfg.VAPIDSubject,
				TTL:             60,
			})
		}()
	}
}
