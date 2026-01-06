// notifier.go
package notifier

import "context"

// Notifier defines a contract for sending notifications.
// Any implementation (Slack, Email, PagerDuty) must satisfy this.
type Notifier interface {
	Send(ctx context.Context, msg Message) error
	Name() string
}

// Message represents the notification message structure.
// message.go

type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityWarning  Severity = "WARNING"
	SeverityCritical Severity = "CRITICAL"
)

type Message struct {
	Title    string
	Body     string
	Severity Severity
}
