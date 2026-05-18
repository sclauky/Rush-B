package utils

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func emailConfig() EmailConfig {
	return EmailConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
		From:     os.Getenv("SMTP_FROM"),
	}
}

func Send2FACode(toEmail, toName, code string) error {
	cfg := emailConfig()

	if cfg.Host == "" || cfg.Username == "" {
		fmt.Printf("[2FA DEV] Code pour %s : %s\n", toEmail, code)
		return nil
	}

	tmplBytes, err := os.ReadFile("backend/forum/auth/templates/email_2fa.html")
	if err != nil {
		return fmt.Errorf("lecture template email: %w", err)
	}

	body := strings.ReplaceAll(string(tmplBytes), "{{NAME}}", toName)
	body = strings.ReplaceAll(body, "{{CODE}}", code)

	msg := buildMIMEMessage(cfg.From, toEmail, "RUSH-B — Ton code de connexion", body)

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	return smtp.SendMail(cfg.Host+":"+cfg.Port, auth, cfg.From, []string{toEmail}, []byte(msg))
}

func buildMIMEMessage(from, to, subject, htmlBody string) string {
	var sb strings.Builder
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString(fmt.Sprintf("From: %s\r\n", from))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", to))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	sb.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(htmlBody)
	return sb.String()
}