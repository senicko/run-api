package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Config represents sandbox's container configuration.
type Config struct {
	Language string `json:"language"` // Docker image that will be run inside sandbox's container.
}

// ExecRequest represents a code exec request.
type ExecRequest struct {
	Config Config `json:"config"`
	Stdin  string `json:"stdin"`
	Files  []struct {
		Name string `json:"name"`
		Body string `json:"body"`
	} `json:"files"`
}

// ExecResponse represents a code exec result.
type ExecResponse struct {
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ExecError string `json:"execError"`
	ExitCode  int    `json:"exitCode"`
}

// Sandbox stores information about created sandbox and it's container.
type Sandbox struct {
	ContainerID string
	cli         *client.Client
}

// StartSandbox creates a new dcker container that can be used as a sandbox.
func CreateSandbox(ctx context.Context, cli *client.Client, config *Config) (*Sandbox, error) {
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:           "code-runner/" + config.Language,
		OpenStdin:       true,
		NetworkDisabled: true,
	}, nil, nil, nil, "")

	if err != nil {
		return nil, fmt.Errorf("sandbox.Sandbox.StartSandbox: failed to create the container: %w", err)
	}

	return &Sandbox{
		ContainerID: resp.ID,
		cli:         cli,
	}, nil
}

// Exec executes submitted code inside of sandbox's container.
func (s *Sandbox) Exec(ctx context.Context, payload *ExecRequest) (*ExecResponse, error) {
	if err := s.start(ctx); err != nil {
		return nil, fmt.Errorf("sandbox.Sandbox.Exec: failed to start the sandbox: %w", err)
	}

	conn, err := s.attach(ctx)
	if err != nil {
		return nil, fmt.Errorf("sandbox.Sandbox.Exec: failed to attach to the sandbox: %w", err)
	}
	defer conn.Close()

	if err := write(conn, payload); err != nil {
		return nil, fmt.Errorf("sandbox.Sandbox.Exec: failed to write the config: %w", err)
	}

	stdout, stderr, err := read(conn)
	if err != nil {
		return nil, fmt.Errorf("sandbox.Sandbox.Exec: failed to read the response: %w", err)
	}

	if stderr.Len() != 0 {
		return nil, fmt.Errorf("sanbox.Sandbox.Exec: sandbox exec failed: %s", stderr.String())
	}

	var response ExecResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return nil, fmt.Errorf("sandbox.Sandbox.Exec: failed to unmarshal sandbox's response: %w", err)
	}

	return &response, nil
}

// Kill stops and removes sandbox's container.
func (s *Sandbox) Kill() {
	ctx := context.Background()
	s.cli.ContainerStop(ctx, s.ContainerID, nil)
	s.cli.ContainerRemove(ctx, s.ContainerID, types.ContainerRemoveOptions{})
}

// start starts the sandbox's docker container.
func (s *Sandbox) start(ctx context.Context) error {
	if err := s.cli.ContainerStart(ctx, s.ContainerID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("sandbox.Sandbox.StartSandbox: failed to start the container: %w", err)
	}

	return nil
}

// attach attaches a connection to the sanbox's container.
func (s *Sandbox) attach(ctx context.Context) (types.HijackedResponse, error) {
	attach, err := s.cli.ContainerAttach(ctx, s.ContainerID, types.ContainerAttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	})
	if err != nil {
		return types.HijackedResponse{}, fmt.Errorf("sandbox.Sandbox.StartSandbox: failed to attach to the container: %w", err)
	}

	return attach, nil
}

// write writes to hijacked response net.Conn.
func write(conn types.HijackedResponse, payload any) error {
	payloadJson, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("sandbox.write: failed to marshal: %w", err)
	}

	if _, err := conn.Conn.Write(append(payloadJson, '\n')); err != nil {
		return fmt.Errorf("sandbox.write: failed to write: %w", err)
	}

	return nil
}

// read reads stdout and stderr from hijacked response.
func read(conn types.HijackedResponse) (bytes.Buffer, bytes.Buffer, error) {
	var stdout, stderr bytes.Buffer

	if _, err := stdcopy.StdCopy(&stdout, &stderr, conn.Reader); err != nil {
		return bytes.Buffer{}, bytes.Buffer{}, fmt.Errorf("sandbox.read: failed to read std: %w", err)
	}

	return stdout, stderr, nil
}
