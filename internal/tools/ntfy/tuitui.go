package ntfy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"saurfang/internal/models/notify"
)

// DingTalkNotification 推推通知
type TuiTuiNotification struct{}
type TuiTuiConfig struct {
	URL    string `json:"url"`
	Token  string `json:"token"`
	Secret string `json:"secret"`
	Team   string `json:"team"`
}

// Send 发送推推通知
func (t *TuiTuiNotification) Send(subject string, message string, cnf *notify.NotifyConfig) error {
	var tuitui TuiTuiConfig
	if err := json.Unmarshal([]byte(cnf.Config), &tuitui); err != nil {
		slog.Error("unmarshal tuitui notification config error", "error", err)
		return err
	}
	alarmData := url.Values{}
	alarmData.Set("teams", tuitui.Team)
	alarmData.Set("title", subject)
	alarmData.Set("app_content", message)
	req, err := http.NewRequest("POST", tuitui.URL, bytes.NewBufferString(alarmData.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(tuitui.Token, tuitui.Secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("tuitui notification error, status code: %d, body: %s", res.StatusCode, body)
	}
	return nil
}
