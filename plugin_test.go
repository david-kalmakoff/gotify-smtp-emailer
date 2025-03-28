package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/david-kalmakoff/gotify-smtp-emailer/testlib"
	"github.com/gotify/plugin-api"
	"github.com/stretchr/testify/require"
)

func TestAPICompatibility(t *testing.T) {
	require.Implements(t, (*plugin.Plugin)(nil), new(Plugin))
	// Add other interfaces you intend to implement here
}

func TestAPI(t *testing.T) {
	s := setup(t)
	defer stop(t, s)

	cfg := baseConfig
	cfg.Hostname = s.Url
	cfg.Token = s.Token
	cfg.Environment = "development"
	cfg.Smtp = Smtp{
		Host:     "localhost",
		Port:     s.MailhogPort,
		Insecure: true,
		Username: "from@email.com",
		Password: toPtr("password"),
		Subject:  toPtr("Test Subject"),
		From: EmailFrom{
			Email: toPtr("from@email.com"),
			Name:  toPtr("Gotify SMTP Emailer"),
		},
		ToEmails: []string{"to@email.com"},
	}

	p := new(Plugin)
	err := p.ValidateAndSetConfig(&cfg)
	require.NoError(t, err)
	t.Logf("\tP\tshould update config")

	err = p.Enable()
	require.NoError(t, err)
	t.Logf("\tP\tshould enable plugin")

	// Give plugin time to connect to websocket
	time.Sleep(1500 * time.Millisecond)

	err = p.Disable()
	require.NoError(t, err)
	t.Logf("\tP\tshould disable plugin")
}

func setup(t *testing.T) *testlib.DockerService {
	// Start docker services
	ctx := context.Background()
	filename := fmt.Sprintf("gotify-smtp-emailer-linux-amd64%s.so", os.Getenv("FILE_SUFFIX"))
	binPath, err := filepath.Abs(filepath.Join("build", filename))
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
