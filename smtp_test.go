package main

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func toPtr[T any](in T) *T {
	return &in
}

func TestSmtpSend(t *testing.T) {
	s := setup(t)
	defer stop(t, s)

	tests := []struct {
		name    string
		domain  string
		mailbox string
		smtp    Smtp
	}{
		{
			name:    "should send email from username",
			domain:  "email.com",
			mailbox: "from",
			smtp: Smtp{
				Host:     "localhost",
				Port:     s.MailhogPort,
				Insecure: true,
				Username: "from@email.com",
				Password: toPtr("password"),
				Subject:  toPtr("Test Subject"),
				From:     EmailFrom{},
				ToEmails: []string{"to@email.com"},
			},
		},
		{
			name:    "should send email from from",
			domain:  "email.com",
			mailbox: "from",
			smtp: Smtp{
				Host:     "localhost",
				Port:     s.MailhogPort,
				Insecure: true,
				Username: "username@email.com",
				Password: toPtr("password"),
				Subject:  toPtr("Test Subject"),
				From: EmailFrom{
					Email: toPtr("from@email.com"),
					Name:  toPtr("Gotify SMTP Emailer"),
				},
				ToEmails: []string{"to@email.com"},
			},
		},
	}

	for i, tt := range tests {
		test := func(t *testing.T) {
			t.Logf("when testing #%d: %s", i, tt.name)

			err := tt.smtp.isValid()
			require.NoError(t, err)

			err = tt.smtp.Send("test title", "test message")
			require.NoError(t, err)

			// Get client token
			res, err := http.Get(s.MailhogUrl + "api/v2/messages")
			require.NoError(t, err)
			require.Equal(t, res.StatusCode, http.StatusOK)

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			var data MailhogRes
			err = json.Unmarshal(body, &data)
			require.NoError(t, err)

			want := MailhogItem{
				From: MailhogMail{
					Domain:  tt.domain,
					Mailbox: tt.mailbox,
				},
			}

			require.Equal(t, want, data.Items[0])
		}

		t.Run(tt.name, test)
	}
}

type MailhogRes struct {
	Count int
	Items []MailhogItem
}

type MailhogItem struct {
	From struct {
		Domain  string
		Mailbox string
	}
}

type MailhogMail struct {
	Domain  string
	Mailbox string
}
