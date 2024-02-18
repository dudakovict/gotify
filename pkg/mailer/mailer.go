// Package mailer provides an interface and a Gmail implementation for sending emails.
package mailer

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

const (
	smtpAuthAddress   = "smtp.gmail.com"
	smtpServerAddress = "smtp.gmail.com:587"
)

// Mailer is the interface that defines the method for sending emails.
type Mailer interface {
	SendEmail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachFiles []string,
	) error
}

// GmailMailer is an implementation of the Mailer interface that
// sends emails using Gmail's SMTP server.
type GmailMailer struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

// NewGmailMailer creates a new GmailMailer instance.
func NewGmailMailer(name string, fromEmailAddress string, fromEmailPassword string) Mailer {
	return &GmailMailer{
		name:              name,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
	}
}

func (mailer *GmailMailer) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
) error {
	e := &email.Email{
		From:    fmt.Sprintf("%s <%s>", mailer.name, mailer.fromEmailAddress),
		Subject: subject,
		HTML:    []byte(content),
		To:      to,
		Cc:      cc,
		Bcc:     bcc,
	}

	for _, f := range attachFiles {
		_, err := e.AttachFile(f)
		if err != nil {
			return fmt.Errorf("failed to attach file %s: %w", f, err)
		}
	}

	smtpAuth := smtp.PlainAuth("", mailer.fromEmailAddress, mailer.fromEmailPassword, smtpAuthAddress)

	return e.Send(smtpServerAddress, smtpAuth)
}
