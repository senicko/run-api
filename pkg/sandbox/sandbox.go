package sandbox

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Request represents a sandbox run request.
type Request struct {
	Language string `json:"language"`
	Files    []struct {
		Name string `json:"name"`
		Body string `json:"body"`
	} `json:"files"`
}

// Response represents a response returned by sandbox.
type Response struct {
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ExecError string `json:"execError"`
	ExitCode  int    `json:"exitCode"`
}

// Run runs run request.
func Run(ctx context.Context, cli *client.Client, runRequest Request) (*Response, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	sandboxID, err := prepareSandbox(ctx, cli, runRequest.Language)
	if err != nil {
		return nil, fmt.Errorf("failed to start the sandbox: %w", err)
	}

	defer removeSandbox(cli, sandboxID)

	result, err := runInSandbox(ctx, cli, sandboxID, runRequest)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// prepareSandbox creates a new sandbox.
func prepareSandbox(ctx context.Context, cli *client.Client, language string) (string, error) {
	image := "code-runner/" + language

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:           image,
		OpenStdin:       true,
		NetworkDisabled: true,
	}, nil, nil, nil, "")
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// runInSandbox runs the code inside sandbox's docker container.
func runInSandbox(ctx context.Context, cli *client.Client, sandboxID string, runRequest Request) (*Response, error) {
	if err := cli.ContainerStart(ctx, sandboxID, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start the sandbox: %w", err)
	}

	attach, err := cli.ContainerAttach(ctx, sandboxID, types.ContainerAttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to the container: %w", err)
	}
	defer attach.Close()

	runRequestJson, err := json.Marshal(runRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the request: %w", err)
	}

	if err := writeToStdin(ctx, attach.Conn, runRequestJson); err != nil {
		return nil, fmt.Errorf("failed to write to the stdin: %w", err)
	}

	stdout, stderr, err := readStd(attach.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read the stdout: %w", err)
	}

	if len(stderr) != 0 {
		return nil, fmt.Errorf("run-error: %s", string(stderr))
	}

	var result Response
	if err := json.Unmarshal(stdout, &result); err != nil {
		fmt.Println(string(stdout))
		return nil, fmt.Errorf("failed to unmarshall: %w", err)
	}

	return &result, nil
}

// removeSandbox removes sandbox's container from the docker host.
func removeSandbox(cli *client.Client, sandboxID string) {
	ctx := context.Background()

	// These shouldn't return error.
	cli.ContainerStop(ctx, sandboxID, nil)
	cli.ContainerRemove(ctx, sandboxID, types.ContainerRemoveOptions{})
}

// writeToStdin writes to sandbox's container stdin.
func writeToStdin(ctx context.Context, writer net.Conn, payload []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if payload[len(payload)-1] != '\n' {
		payload = append(payload, '\n')
	}

	if _, err := writer.Write(payload); err != nil {
		return err
	}

	return nil
}

// readStd reads stdout from container's hijacked response reader.
func readStd(reader *bufio.Reader) ([]byte, []byte, error) {
	var stdout, stderr bytes.Buffer

	if _, err := stdcopy.StdCopy(&stdout, &stderr, reader); err != nil {
		return nil, nil, err
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}
