package ntfy

import (
	"context"
	"encoding/json"
	"log/slog"
	"saurfang/internal/models/notify"
	"time"

	"github.com/nikoksr/notify/service/dingding"
)

// DingTalkNotification 钉钉通知
type DingTalkNotification struct{}
type DIngTalkConfig struct {
	Token  string `json:"token"`
	Secret string `json:"secret"`
}

// Send 发送钉钉通知
func (n *DingTalkNotification) Send(subject string, message string, cnf *notify.NotifyConfig) error {
	var dingtalk DIngTalkConfig
	if err := json.Unmarshal([]byte(cnf.Config), &dingtalk); err != nil {
		slog.Error("unmarshal dingtalk notification config error", "error", err)
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cfg := dingding.Config{
		Token:  dingtalk.Token,
		Secret: dingtalk.Secret,
	}
	dingdingSvc := dingding.New(&cfg)
	err := dingdingSvc.Send(ctx, subject, message)
	if err != nil {
		return err
	}
	return nil
}
