package ntfy

import (
	"context"
	"encoding/json"
	"log/slog"
	"saurfang/internal/models/notify"
	"time"

	gonotify "github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/lark"
)

// LarkNotification 飞书通知
type LarkNotification struct{}
type LarkConfig struct {
	Webhook string
}

func (l *LarkNotification) Send(subject, message string, cnf *notify.NotifyConfig) error {
	var larkConfig LarkConfig
	if err := json.Unmarshal([]byte(cnf.Config), &larkConfig); err != nil {
		slog.Error("unmarshal lark notification config error", "error", err)
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	larkwebhookSvc := lark.NewWebhookService(larkConfig.Webhook)
	notifier := gonotify.New()
	notifier.UseServices(larkwebhookSvc)
	if err := notifier.Send(ctx, subject, message); err != nil {
		return err
	}
	return nil
}
