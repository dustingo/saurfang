package ntfy

import (
	"context"
	"time"

	gonotify "github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/lark"
)

// LarkNotification 飞书通知
type LarkNotification struct {
	Webhook string
}

func (l *LarkNotification) Send(subject, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	larkwebhookSvc := lark.NewWebhookService(l.Webhook)
	notifier := gonotify.New()
	notifier.UseServices(larkwebhookSvc)
	if err := notifier.Send(ctx, subject, message); err != nil {
		return err
	}
	return nil
}
