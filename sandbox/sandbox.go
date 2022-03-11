package sandbox

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Request struct {
	Language string `json:"language"`
	Files    []struct {
		Name string `json:"name"`
		Body string `json:"body"`
	} `json:"files"`
}

type Response struct {
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ExecError string `json:"execError"`
	ExitCode  int    `json:"exitCode"`
}

func PrepareSandbox(ctx context.Context, cli *client.Client, language string) (string, error) {
	image := "code-runner/" + language

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:           image,
		OpenStdin:       true,
		NetworkDisabled: true,
	}, nil, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("error: failed to create a container: %w", err)
	}

	return resp.ID, nil
}

func Run(ctx context.Context, cli *client.Client, sID string, runRequest Request) (*Response, error) {
	if err := cli.ContainerStart(ctx, sID, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start the sandbox: %w", err)
	}

	attach, err := cli.ContainerAttach(ctx, sID, types.ContainerAttachOptions{
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

	if err := writeStd(ctx, attach.Conn, runRequestJson); err != nil {
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

func KillSandbox(cli *client.Client, sandboxID string) {
	ctx := context.Background()
	cli.ContainerStop(ctx, sandboxID, nil)
	cli.ContainerRemove(ctx, sandboxID, types.ContainerRemoveOptions{})
}
