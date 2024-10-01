package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/david-kalmakoff/gotify-smtp-emailer/testlib"
	"github.com/gotify/plugin-api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPICompatibility(t *testing.T) {
	assert.Implements(t, (*plugin.Plugin)(nil), new(Plugin))
	// Add other interfaces you intend to implement here
}

func TestAPI(t *testing.T) {
	s := setup(t)
	defer stop(t, s)

	cfg := baseConfig
	cfg.Hostname = s.Url
	cfg.Token = s.Token
	cfg.Smtp = Smtp{
		Host:      "localhost",
		Port:      s.MailhogPort,
		FromEmail: "from@email.com",
		Password:  "password",
		ToEmails:  []string{"to@email.com"},
		Subject:   "Test Subject",
	}

	p := new(Plugin)
	err := p.ValidateAndSetConfig(&cfg)
	assert.NoError(t, err)

	err = p.Enable()
	assert.NoError(t, err)

	err = p.Disable()
	assert.NoError(t, err)
}

func setup(t *testing.T) *testlib.DockerService {
	// Start docker services
	ctx := context.Background()
	binPath, err := filepath.Abs(filepath.Join("build", "gotify-smtp-emailer-linux-amd64.so"))
	require.NoError(t, err)
	s, err := testlib.NewDockerService(ctx, binPath)
	require.NoError(t, err)

	return s
}

func stop(t *testing.T, s *testlib.DockerService) {
	// Stop docker services
	ctx := context.Background()
	err := s.Stop(ctx)
	require.NoError(t, err)
}
