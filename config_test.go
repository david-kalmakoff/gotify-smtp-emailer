package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var baseConfig = Config{
	Hostname: "http://localhost",
	Token:    "token",
	Smtp: Smtp{
		Host:      "smtp.host.com",
		Port:      587,
		FromEmail: "from@email.com",
		Password:  "password",
		ToEmails:  []string{"to@email.com"},
		Subject:   "Test Subject",
		Insecure:  false,
	},
	Environment: "production",
}

func TestConfigIsValid(t *testing.T) {
	s := setup(t)
	defer stop(t, s)

	tests := []struct {
		name   string
		config func() Config
		pass   bool
	}{
		{
			name: "should create valid config",
			config: func() Config {
				cfg := baseConfig
				cfg.Hostname = s.Url
				cfg.Token = s.Token
				return cfg
			},
			pass: true,
		},
		{
			name: "should not create invalid config",
			config: func() Config {
				cfg := baseConfig
				return cfg
			},
			pass: false,
		},
	}

	for i, tt := range tests {
		test := func(t *testing.T) {
			t.Logf("When testing #%d: %s", i, tt.name)
			cfg := tt.config()
			err := cfg.IsValid()
			if tt.pass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		}

		t.Run(tt.name, test)
	}
}
