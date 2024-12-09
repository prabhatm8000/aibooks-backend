package utils

import (
	"example/aibooks-backend/errorHandling"
	"net/smtp"
	"os"
)

func Ternary[T any](condition bool, trueValue T, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}

func sendEmail(recipient, subject, body string) error {
	sender := os.Getenv("EMAIL")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	message := []byte("Subject: " + subject + "\n\n" + body)

	auth := smtp.PlainAuth("", sender, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, sender, []string{recipient}, message)
	if err != nil {
		return errorHandling.NewAPIError(500, sendEmail, err.Error())
	}
	return nil
}

func SendOtpEmail(recipient, opt string) error {
	err := sendEmail(recipient, "OTP", "Your OTP is "+opt)
	if err == nil {
		return nil
	}
	return errorHandling.NewAPIError(500, SendOtpEmail, err.Error())
}
