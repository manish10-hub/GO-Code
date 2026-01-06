package notifier

import (
	"context"
	"fmt"
	"time"
)

type SlackNotifier struct {
	WebhookURL string
	Timeout    time.Duration
}

func (s *SlackNotifier) Name() string {
	return "slack"
}

func (s *SlackNotifier) Send(ctx context.Context, msg Message) error {
	// In real code, this would be an HTTP call
	select {
	case <-time.After(100 * time.Millisecond):
		fmt.Printf("[Slack] %s: %s\n", msg.Title, msg.Body)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
