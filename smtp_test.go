package main

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSmtpSend(t *testing.T) {
	s := setup(t)
	defer stop(t, s)

	smtp := Smtp{
		Host:      "localhost",
		Port:      s.MailhogPort,
		FromEmail: "from@email.com",
		Password:  "password",
		ToEmails:  []string{"to@email.com"},
		Subject:   "Test Subject",
	}

	err := smtp.isValid()
	require.NoError(t, err)

	err = smtp.Send("test title", "test message")
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

	want := MailhogRes{
		Count: 1,
		Items: []MailhogItem{
			{
				From: MailhogMail{
					Domain:  "email.com",
					Mailbox: "from",
				},
			},
		},
	}

	require.Equal(t, want, data)
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
