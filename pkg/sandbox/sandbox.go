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
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

func Run(ctx context.Context, cli *client.Client, runRequest Request) (*Response, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	sandboxID, err := prepareSandbox(ctx, cli)
	if err != nil {
		return nil, fmt.Errorf("could not start the sandbox: %w", err)
	}

	// FIXME: should handle the error
	defer removeSandbox(cli, sandboxID)

	result, err := runInSandbox(ctx, cli, sandboxID, runRequest)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// prepareSandbox creates a new sandbox.
func prepareSandbox(ctx context.Context, cli *client.Client) (string, error) {
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:     "bee/golang",
		OpenStdin: true,
	}, nil, nil, nil, "")
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// runInSandbox runs the code inside sandbox's docker container.
func runInSandbox(ctx context.Context, cli *client.Client, sandboxID string, runRequest Request) (*Response, error) {
	if err := cli.ContainerStart(ctx, sandboxID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	attach, err := cli.ContainerAttach(ctx, sandboxID, types.ContainerAttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	})
	if err != nil {
		return nil, err
	}
	defer attach.Close()

	runRequestJson, err := json.Marshal(runRequest)
	if err != nil {
		return nil, err
	}

	if err := writeToStdin(ctx, attach.Conn, runRequestJson); err != nil {
		return nil, err
	}

	stdout, stderr, err := readStd(ctx, attach.Reader)
	if err != nil {
		return nil, err
	}

	if len(stderr) != 0 {
		return nil, fmt.Errorf("run-error: %s", string(stderr))
	}

	// TODO: probably rename RunResponse to RunResult
	var result Response
	if err := json.Unmarshal(stdout, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// removeSandbox removes sandbox's container from the docker host.
func removeSandbox(cli *client.Client, sandboxID string) error {
	ctx := context.Background()

	if err := cli.ContainerStop(ctx, sandboxID, nil); err != nil {
		fmt.Println("stopping error: ", err)
		return err
	}

	if err := cli.ContainerRemove(ctx, sandboxID, types.ContainerRemoveOptions{}); err != nil {
		fmt.Println("removing error: ", err)
		return err
	}

	return nil
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
func readStd(ctx context.Context, reader *bufio.Reader) ([]byte, []byte, error) {
	// TODO: this can be implemented much better

	resultChan := make(chan [2][]byte)
	errChan := make(chan error)

	go func() {
		var stdout, stderr bytes.Buffer

		if _, err := stdcopy.StdCopy(&stdout, &stderr, reader); err != nil {
			errChan <- err
		}

		resultChan <- [2][]byte{
			stdout.Bytes(),
			stderr.Bytes(),
		}
	}()

	select {
	case <-ctx.Done():
		return nil, nil, fmt.Errorf("timeout")
	case err := <-errChan:
		return nil, nil, err
	case result := <-resultChan:
		return result[0], result[1], nil
	}
}
