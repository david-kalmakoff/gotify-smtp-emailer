package testlib_test

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/david-kalmakoff/gotify-smtp-mailer/testlib"
	"github.com/stretchr/testify/require"
)

func TestWithGotify(t *testing.T) {
	s := setup(t)
	defer stop(t, s)

	// Test Gotify endpoint
	req, err := http.NewRequest(http.MethodGet, s.UrlAuth+"plugin", nil)
	require.NoError(t, err)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, res.StatusCode, http.StatusOK)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Contains(t, string(body), "Gotify SMTP Emailer")
}

func setup(t *testing.T) *testlib.DockerService {
	// Start docker services
	ctx := context.Background()
	binPath, err := filepath.Abs(filepath.Join("..", "build", "gotify-smtp-emailer-linux-amd64.so"))
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
