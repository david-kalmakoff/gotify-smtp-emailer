package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Config struct {
	Hostname    string // This will be local because they are running on same machine
	Token       string //Token from client needed for ws connection
	Smtp        Smtp
	Environment string
}

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

	// test websocket connection
	conn, err := c.getWSConnection()
	if err != nil {
		return fmt.Errorf("could not get ws connection: %w", err)
	}
	defer conn.Close()

	// validate smtp
	err = c.Smtp.isValid()
	if err != nil {
		return fmt.Errorf("smtp is invalid: %w", err)
	}

	return nil
}

func (c *Config) getWSConnection() (*websocket.Conn, error) {
	count := 0
	for {
		count++
		uri := fmt.Sprintf("%s/stream?token=%s", c.Hostname, c.Token)
		ws, _, err := websocket.DefaultDialer.Dial(uri, nil)
		if err == nil {
			return ws, nil
		}
		if count > 3 {
			return nil, fmt.Errorf("Cannot connect to websocket %q: %w", uri, err)

		}
		time.Sleep(500 * time.Millisecond)
	}
}

// ============================================================================

func (c *Plugin) DefaultConfig() interface{} {
	return &Config{
		Hostname: "ws://localhost",
		Token:    "",
		Smtp: Smtp{
			Host:      "smtp.example.com",
			Port:      587,
			FromEmail: "from@email.com",
			ToEmails:  []string{"to@email.com"},
			Password:  "password",
			Subject:   "",
		},
		Environment: "production",
	}
}

// ValidateAndSetConfig runs when the user saves the config
func (c *Plugin) ValidateAndSetConfig(in interface{}) error {
	config, ok := in.(*Config)
	if !ok {
		return errors.New("invalid config")
	}

	err := config.IsValid()
	if err != nil {
		return fmt.Errorf("config is invalid: %w", err)
	}

	c.config = config

	return nil
}
