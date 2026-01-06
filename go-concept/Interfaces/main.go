package main

import (
	"context"
	"time"

	"example.com/notifier"
)

func main() {
	slack := &notifier.SlackNotifier{
		WebhookURL: "https://hooks.slack.com/xxx",
		Timeout:    2 * time.Second,
	}

	email := &notifier.EmailNotifier{
		SMTPServer: "smtp.company.com",
		From:       "alerts@company.com",
	}

	alertManager := notifier.NewAlertManager(slack, email)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	alertManager.NotifyAll(ctx, notifier.Message{
		Title:    "Disk Pressure",
		Body:     "Node xyz is running out of disk space",
		Severity: notifier.SeverityCritical,
	})
}
