package ntfy

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"saurfang/internal/models/notify"
	"strings"
)

// HTTPNotification HTTP通知
type HTTPNotification struct{}
type HttpConfig struct {
	URL         string      `json:"URL"`
	Header      http.Header `json:"Header"`
	ContentType string      `json:"ContentType"`
}

func (n *HTTPNotification) Send(subject string, message string, cnf *notify.NotifyConfig) error {
	var hc HttpConfig
	err := json.Unmarshal([]byte(cnf.Config), &hc)
	if err != nil {
		slog.Error("unmarshal http config error", "error", err)
		return err
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", hc.URL, strings.NewReader(message))
	if err != nil {
		slog.Error("new http request error", "error", err)
		return err
	}
	req.Header.Set("Content-Type", hc.ContentType)
	for k, v := range hc.Header {
		req.Header.Set(k, v[0])
	}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("do http request error", "error", err)
		return err
	}
	defer resp.Body.Close()
	slog.Info("send http notification", "url", hc.URL, "header", hc.Header, "contenttype", hc.ContentType)

	return nil
}
