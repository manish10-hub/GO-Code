package notifier

import (
	"context"
	"fmt"
)

type AlertManager struct {
	notifiers []Notifier
}

func NewAlertManager(notifiers ...Notifier) *AlertManager {
	return &AlertManager{
		notifiers: notifiers,
	}
}

func (a *AlertManager) NotifyAll(ctx context.Context, msg Message) {
	for _, n := range a.notifiers {
		if err := n.Send(ctx, msg); err != nil {
			fmt.Printf("Notifier %s failed: %v\n", n.Name(), err)
		}
	}
}
