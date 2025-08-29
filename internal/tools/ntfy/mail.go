package ntfy

import (
	"encoding/json"
	"log/slog"
	"os"
	"saurfang/internal/models/notify"
	"strconv"

	"github.com/Shopify/gomail"
)

// EmailNotification 邮件通知
type EmailNotification struct{}
type EmailConfig struct {
	To []string
}

func (n *EmailNotification) Send(subject string, message string, cnf *notify.NotifyConfig) error {
	var email EmailConfig
	if err := json.Unmarshal([]byte(cnf.Config), &email); err != nil {
		slog.Error("unmarshal email notification config error", "error", err)
		return err
	}
	userName := os.Getenv("EMAIL_USERNAME")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("EMAIL_SMTP_HOST")
	smtpPort, _ := strconv.Atoi(os.Getenv("EMAIL_SMTP_PORT"))
	// 创建邮件服务
	mailer := gomail.NewDialer(smtpHost, smtpPort, userName, password)
	mail := gomail.NewMessage()
	mail.SetHeader("From", userName)
	mail.SetHeader("To", email.To...)
	mail.SetHeader("Subject", subject)
	mail.SetBody("text/plain", message)
	// 发送邮件
	if err := mailer.DialAndSend(mail); err != nil {
		return err
	}
	return nil
}
