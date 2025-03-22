package testlib

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

type DockerService struct {
	GotifyContainer  testcontainers.Container
	MailhogContainer testcontainers.Container
	Url              string
	UrlAuth          string
	Token            string
	MailhogPort      int
	MailhogUrl       string
	Network          *testcontainers.DockerNetwork
}

func NewDockerService(ctx context.Context, binPath string, envs ...string) (*DockerService, error) {
	s := DockerService{}

	var err error
	s.Network, err = network.New(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not create network")
	}
	networkName := s.Network.Name

	envVars := make(map[string]string)
	i := 0
	for i < len(envs) {
		envVars[envs[i]] = envs[i+1]
		i = i + 2
	}

	filename := fmt.Sprintf("gotify-smtp-emailer-linux-amd64%s.so", os.Getenv("FILE_SUFFIX"))
	version := strings.ReplaceAll(os.Getenv("GOTIFY_VERSION"), "v", "")
	imageName := fmt.Sprintf("gotify/server:%s", version)
	req := testcontainers.ContainerRequest{
		Image:        imageName,
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForLog("Started listening for plain connection on tcp [::]:80"),
		Networks:     []string{networkName},
		Env:          envVars,
		Mounts: testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericBindMountSource{
					HostPath: binPath,
				},
				Target: testcontainers.ContainerMountTarget("/app/data/plugins/" + filename),
			},
		},
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{&testcontainers.StdoutLogConsumer{}},
		},
	}
	s.GotifyContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create container: %w", err)
	}

	ip, err := s.GotifyContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get container host: %w", err)
	}

	mappedPort, err := s.GotifyContainer.MappedPort(ctx, "80/tcp")
	if err != nil {
		return nil, fmt.Errorf("could not get container port: %w", err)
	}

	// Add basic auth to url
	s.UrlAuth = fmt.Sprintf("http://admin:admin@%s:%s/", ip, mappedPort.Port())
	s.Url = fmt.Sprintf("http://%s:%s/", ip, mappedPort.Port())

	// Get client token
	httpReq, err := http.NewRequest(http.MethodPost, s.UrlAuth+"client", bytes.NewBuffer([]byte(`{"name":"client"}`)))
	if err != nil {
		return nil, fmt.Errorf("could not make client request: %w", err)
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Accept", "application/json")
	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("could not do request: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read body: %w", err)
	}

	var httpRes struct {
		Id    int
		Name  string
		Token string
	}
	err = json.Unmarshal(body, &httpRes)
	if err != nil {
		return nil, fmt.Errorf("could unmarshal body: %w", err)
	}
	if httpRes.Token == "" {
		return nil, errors.New("no token found")
	}

	s.Token = httpRes.Token

	// setup mailhog
	req = testcontainers.ContainerRequest{
		Hostname:     "mailhog",
		Image:        "mailhog/mailhog:v1.0.1",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor:   wait.ForLog("Creating API v2 with WebPath:"),
		Networks:     []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"mailhog"},
		},
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{&testcontainers.StdoutLogConsumer{}},
		},
	}
	s.MailhogContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create mailhog container: %w", err)
	}

	ip, err = s.MailhogContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get mailhog host: %w", err)
	}

	mappedPort, err = s.MailhogContainer.MappedPort(ctx, "8025/tcp")
	if err != nil {
		return nil, fmt.Errorf("could not get mailhog web port: %w", err)
	}
	s.MailhogUrl = fmt.Sprintf("http://%s:%s/", ip, mappedPort.Port())

	mappedPort, err = s.MailhogContainer.MappedPort(ctx, "1025/tcp")
	if err != nil {
		return nil, fmt.Errorf("could not get mailhog smtp port: %w", err)
	}
	s.MailhogPort = mappedPort.Int()

	fmt.Printf("Gotify available at: %s\n", s.Url)
	fmt.Println("Gotify username=admin, password=admin")
	fmt.Printf("Mailhog available at: %s\n", s.MailhogUrl)
	fmt.Printf("Mailhog port: %d\n", s.MailhogPort)

	return &s, nil
}

func (s *DockerService) Stop(ctx context.Context) error {
	err := s.GotifyContainer.Terminate(ctx)
	if err != nil {
		return fmt.Errorf("could not terminate: %w", err)
	}
	err = s.MailhogContainer.Terminate(ctx)
	if err != nil {
		return fmt.Errorf("could not terminate mailhog: %w", err)
	}
	err = s.Network.Remove(ctx)
	if err != nil {
		return fmt.Errorf("could not remove network: %w", err)
	}

	return nil
}
