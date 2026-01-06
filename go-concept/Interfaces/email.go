package notifier

import (
	"context"
	"fmt"
)

type EmailNotifier struct {
	SMTPServer string
	From       string
}

func (e *EmailNotifier) Name() string {
	return "email"
}

func (e *EmailNotifier) Send(ctx context.Context, msg Message) error {
	fmt.Printf("[Email] %s: %s\n", msg.Title, msg.Body)
	return nil
}
