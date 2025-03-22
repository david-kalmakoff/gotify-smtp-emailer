package main

import (
	"bytes"
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
	Host     string
	Port     int
	Insecure bool
	Username string
	Password *string // Optional: if empty no SMTP auth is used
	Subject  *string // Optional: included subject string
	From     EmailFrom
	ToEmails []string
}

type EmailFrom struct {
	Email *string // Optional: defaults to send from SMTP username
	Name  *string // Optional: name included in email from
}

// isValid is used to validate the Smtp configuration
func (s *Smtp) isValid() error {
	if s.Host == "" {
		return errors.New("the smtp host is not valid")
	}
	if s.Port == 0 {
		return errors.New("the smtp port is not valid")
	}
	if s.Username == "" {
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
	if s.Subject != nil {
		subject = fmt.Sprintf("%s: %s", *s.Subject, title)
	}

	text := "<div>"
	text += "<h3>"
	text += title
	text += "</h3>"
	text += "<p>"
	text += message
	text += "</p>"
	text += "</div>"

	var content bytes.Buffer
	fromEmail := s.Username
	if s.From.Email != nil {
		fromEmail = *s.From.Email
	}

	if s.From.Name != nil {
		content.WriteString(fmt.Sprintf("From: %s <%s>\n", *s.From.Name, fromEmail))
	} else {
		content.WriteString(fmt.Sprintf("From: %s\n", fromEmail))
	}

	content.WriteString(fmt.Sprintf(
		"To: %s\n", strings.Join(s.ToEmails, ", ")))
	content.WriteString(fmt.Sprintf("Subject: %s\n", subject))
	content.WriteString("MIME-version: 1.0;\n")
	content.WriteString("Content-Type: text/html; charset=\"UTF-8\";\n")
	content.WriteString(fmt.Sprintf("Message-ID: <%s>\n\n", messageID))
	content.WriteString(text)

	var auth smtp.Auth
	authType := "nil"
	if s.Password != nil {
		if s.Insecure {
			auth = smtp.CRAMMD5Auth(s.Username, *s.Password)
			authType = "CRAMMD5Auth"
		} else {
			auth = smtp.PlainAuth("", s.Username, *s.Password, s.Host)
			authType = "PlainAuth"
		}
	}

	fmt.Printf("SMTP Emailer: sending with auth='%s'\n", authType)
	uri := fmt.Sprintf("%s:%d", s.Host, s.Port)
	err := smtp.SendMail(uri, auth, fromEmail, s.ToEmails, content.Bytes())
	if err != nil {
		return fmt.Errorf("could not send email: %w", err)
	}

	return nil
}
