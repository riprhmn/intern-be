package service

import (
	"fmt"
	"net/smtp"
	"os"
)

type EmailService struct {
	host     string
	port     string
	username string
	password string
}

func NewEmailService() *EmailService {
	return &EmailService{
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
	}
}

func (s *EmailService) SendResetPasswordEmail(toEmail, resetLink string) error {
	if s.host == "" || s.port == "" || s.username == "" || s.password == "" {
		return fmt.Errorf("SMTP configuration is incomplete")
	}

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	subject := "Reset Password - Magang System"
	body := fmt.Sprintf(`Halo,

Kami menerima permintaan untuk melakukan reset password pada akun Anda.

Silakan klik link berikut untuk membuat password baru:

%s

Link tersebut berlaku selama 15 menit.

Jika Anda tidak meminta reset password, abaikan email ini.

Terima kasih.
Magang System
`, resetLink)

	message := []byte(
		"From: " + s.username + "\r\n" +
			"To: " + toEmail + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" + body,
	)

	return smtp.SendMail(
		s.host+":"+s.port,
		auth,
		s.username,
		[]string{toEmail},
		message,
	)
}
