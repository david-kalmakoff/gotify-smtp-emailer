package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// Smtp represents an SMTP configuration
type Smtp struct {
	Host      string
	Port      int
	FromEmail string
	Password  string // Optional: if empty no SMTP auth is used
	ToEmails  []string
	Subject   string // Optional: included subject string
	Insecure  bool
}

// isValid is used to validate the Smtp configuration
func (s *Smtp) isValid() error {
	if s.Host == "" {
		return errors.New("the smtp host is not valid")
	}
	if s.Port == 0 {
		return errors.New("the smtp port is not valid")
	}
	if s.FromEmail == "" {
		return errors.New("the smtp from email is not valid")
	}
	if len(s.ToEmails) < 1 {
		return errors.New("the smtp to emails are not valid")
	}

	return nil
}

// ============================================================================

// Send is used to send an SMTP email
func (s *Smtp) Send(title, message string) error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	messageID := strconv.FormatInt(r.Int63(), 10) + "@" + s.Host

	subject := title
	if s.Subject != "" {
		subject = fmt.Sprintf("%s: %s", s.Subject, title)
	}

	text := "<div>"
	text += "<h3>"
	text += title
	text += "</h3>"
	text += "<p>"
	text += message
	text += "</p>"
	text += "</div>"

	content := "From: " + s.FromEmail + "\n" +
		"To: " + strings.Join(s.ToEmails, ", ") + "\n" +
		"Subject: " + subject + "\n" +
		"MIME-version: 1.0;\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\n" +
		"Message-ID: <" + messageID + ">\n\n" +
		text

	var auth smtp.Auth
	authType := "nil"
	if s.Password != "" {
		if s.Insecure {
			auth = smtp.CRAMMD5Auth(s.FromEmail, s.Password)
			authType = "CRAMMD5Auth"
		} else {
			auth = smtp.PlainAuth("", s.FromEmail, s.Password, s.Host)
			authType = "PlainAuth"
		}
	}

	fmt.Printf("SMTP Emailer: sending with auth='%s'\n", authType)
	uri := fmt.Sprintf("%s:%d", s.Host, s.Port)
	err := smtp.SendMail(uri, auth, s.FromEmail, s.ToEmails, []byte(content))
	if err != nil {
		return fmt.Errorf("could not send email: %w", err)
	}

	return nil
}
