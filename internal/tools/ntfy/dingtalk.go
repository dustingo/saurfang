package ntfy

import (
	"context"
	"time"

	"github.com/nikoksr/notify/service/dingding"
)

// DingTalkNotification 钉钉通知
type DingTalkNotification struct {
	Token  string `json:"token"`
	Secret string `json:"secret"`
}

// Send 发送钉钉通知
func (n *DingTalkNotification) Send(subject string, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cfg := dingding.Config{
		Token:  n.Token,
		Secret: n.Secret,
	}
	dingdingSvc := dingding.New(&cfg)
	err := dingdingSvc.Send(ctx, subject, message)
	if err != nil {
		return err
	}
	return nil
}
