package ntfy

import "log/slog"

// HTTPNotification HTTP通知
type HTTPNotification struct {
	URL         string
	Header      map[string]string
	ContentType string
}

func (n *HTTPNotification) Send() error {
	slog.Info("send http notification", "url", n.URL, "header", n.Header, "contenttype", n.ContentType)

	return nil
}
