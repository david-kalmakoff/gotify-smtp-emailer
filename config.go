package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

// Config represents the config used for the plugin
type Config struct {
	Hostname string // This will be local because they are running on same machine
	Token    string //Token from client needed for ws connection
	Smtp     Smtp
	// production or development, used for logging and sending messages on a loop
	Environment string
}

// ============================================================================

// IsValid is used to validate the plugin configuration
func (c *Config) IsValid() error {
	// Validate Config
	c.Hostname = strings.TrimSpace(c.Hostname)
	c.Token = strings.TrimSpace(c.Token)

	// make sure there is no trailing /
	c.Hostname = strings.TrimSuffix(c.Hostname, "/")
	// convert http to ws path
	c.Hostname = strings.Replace(c.Hostname, "https://", "wss://", 1)
	c.Hostname = strings.Replace(c.Hostname, "http://", "ws://", 1)

	if len(c.Hostname) < 3 {
		return errors.New("hostname too short")
	}
	if c.Hostname[0] != 'w' || c.Hostname[1] != 's' {
		return errors.New("invalid hostname")
	}
	if c.Hostname == "" {
		return errors.New("the hostname is not valid")
	}
	if c.Token == "" {
		return errors.New("the token is not valid")
	}
	if c.Environment != "production" && c.Environment != "development" {
		return errors.New("the environment is not valid")
	}

	// validate smtp
	err := c.Smtp.isValid()
	if err != nil {
		return fmt.Errorf("smtp is invalid: %w", err)
	}

	return nil
}

// ============================================================================

// DefaultConfig is the default config set for the user
func (c *Plugin) DefaultConfig() interface{} {
	if os.Getenv("ENV") == "development" {
		name := "Gotify SMTP Emailer"
		email := "from@email.com"
		return &Config{
			Hostname: "ws://localhost",
			Token:    "",
			Smtp: Smtp{
				Host:     "mailhog",
				Port:     1025,
				Insecure: true,
				Username: "username@email.com",
				From: EmailFrom{
					Name:  &name,
					Email: &email,
				},
				ToEmails: []string{"to@email.com"},
			},
			Environment: "development",
		}
	}
	return &Config{
		Hostname: "ws://localhost",
		Token:    "",
		Smtp: Smtp{
			Host:     "smtp.example.com",
			Port:     587,
			Insecure: false,
			Username: "username@email.com",
			From:     EmailFrom{},
			ToEmails: []string{"to@email.com"},
		},
		Environment: "production",
	}
}

// ============================================================================

// ValidateAndSetConfig is called when the user saves the config
func (c *Plugin) ValidateAndSetConfig(in any) error {
	config, ok := in.(*Config)
	if !ok {
		return errors.New("invalid config")
	}

	if config.Environment == "development" {
		b, _ := json.MarshalIndent(config, "", " ")
		log.Println("SMTP Emailer: updating config:")
		log.Println(string(b))
	}

	err := config.IsValid()
	if err != nil {
		return fmt.Errorf("config is invalid: %w", err)
	}

	log.Println("SMTP Emailer: updated config")

	c.config = config

	return nil
}
