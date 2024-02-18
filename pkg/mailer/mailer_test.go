package mailer

import (
	"testing"

	"github.com/dudakovict/gotify/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	config, err := config.Load("../../")
	require.NoError(t, err)

	mailer := NewGmailMailer(config.MailerName, config.EmailAddress, config.EmailPassword)

	subject := "A test email"
	content := `
	<h1>Hello World</h1>
	`
	to := []string{"test@example.com"}
	attachFiles := []string{"../../README.md"}

	err = mailer.SendEmail(subject, content, to, nil, nil, attachFiles)
	require.NoError(t, err)
}
