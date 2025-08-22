package ntfy

import "log/slog"

// WeChatNotification 企业微信通知
type WeChatNotification struct {
	AppID          string
	AppSecret      string
	Token          string
	EncodingAESKey string
}

func (n *WeChatNotification) Send() error {
	slog.Info("send wechat notification", "appid", n.AppID, "appsecret", n.AppSecret, "token", n.Token, "encodingaeskey", n.EncodingAESKey)

	return nil
}
