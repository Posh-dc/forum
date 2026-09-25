package backEnd

import (
	"fmt"
	"net/smtp"
	"os"
)

var (
	EmailVerificationSubject = "Verify your Lobaloba X-Forum account\r\n"
	username                 = os.Getenv("SMTP_USERNAME")
	appPassword              = os.Getenv("SMTP_PASSWORD")
	smtpServer               = "smtp.gmail.com"
	smtpPort                 = "587"
	from                     = "lobalobaxforum@gmail.com"
)

func CreateVerificationEmailBody(code string) string {
	return fmt.Sprintf(
		"Hello,\r\n"+
			"\r\n"+
			"Please verify your Lobaloba X-Forum account.\r\n"+
			"\r\n"+
			"Your verification code is: %s\r\n"+
			"\r\n"+
			"This code expires in 10 minutes.\r\n",
		code,
	)
}

type Email struct {
	ID             int64
	RecipientEmail string
	Subject        string
	Body           string
	Attempts       int
}

func SendMail(email Email) error {

	username = os.Getenv("SMTP_USERNAME")
	appPassword = os.Getenv("SMTP_PASSWORD")

	to := []string{email.RecipientEmail}

	message := []byte(
		"From: " + from + "\r\n" +
			"To: " + email.RecipientEmail + "\r\n" +
			"Subject: " + email.Subject + "\r\n" +
			"\r\n" +
			email.Body,
	)

	auth := smtp.PlainAuth(
		"",
		username,
		appPassword,
		smtpServer,
	)

	err := smtp.SendMail(
		smtpServer+":"+smtpPort,
		auth,
		from,
		to,
		message,
	)

	if err != nil {
		return err
	}

	return nil
}
