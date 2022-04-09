package sandbox

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"fmt"

// 	"github.com/docker/docker/api/types"
// 	"github.com/docker/docker/api/types/container"
// 	"github.com/docker/docker/client"
// 	"github.com/docker/docker/pkg/stdcopy"
// )

// type Request struct {
// 	Language string `json:"language"`
// 	Stdin    string `json:"stdin"`
// 	Files    []struct {
// 		Name string `json:"name"`
// 		Body string `json:"body"`
// 	} `json:"files"`
// }

// type Response struct {
// 	Stdout    string `json:"stdout"`
// 	Stderr    string `json:"stderr"`
// 	ExecError string `json:"execError"`
// 	ExitCode  int    `json:"exitCode"`
// }

// // Run is an entry point for code execution.
// // It processes the code execution request and returns the results.
// func Run(ctx context.Context, cli *client.Client, runRequest Request) (*Response, error) {
// 	sID, err := startSandbox(ctx, cli, runRequest.Language)
// 	if err != nil {
//     return nil, fmt.Errorf("sandbox.Run: failed to prepare the sandbox: %w", err)
// 	}

// 	attach, err := cli.ContainerAttach(ctx, sID, types.ContainerAttachOptions{
// 		Stdin:  true,
// 		Stdout: true,
// 		Stderr: true,
// 		Stream: true,
// 	})
// 	if err != nil {
//     return nil, fmt.Errorf("sandbox.Run: failed to attach to the container: %w", err)
// 	}
// 	defer attach.Close()

// 	if err := writeRunRequest(attach, runRequest); err != nil {
//     return nil, fmt.Errorf("sandbox.Run: failed to write to container's stdin: %w", err)
// 	}

// 	stdout, stderr, err := readStd(attach)
// 	if err != nil {
//     return nil, fmt.Errorf("sandbox.Run: failed to read the stdout: %w", err)
// 	}

// 	if len(stderr) != 0 {
//     return nil, fmt.Errorf("sandbox.Run: run-error: %s", string(stderr))
// 	}

// 	var result Response
// 	if err := json.Unmarshal(stdout, &result); err != nil {
//     return nil, fmt.Errorf("sandbox.Run: failed to unmarshall: %w", err)
// 	}

// 	killSandbox(cli, sID)

// 	return &result, nil
// }

// // killSandbox removes the sandbox.
// func killSandbox(cli *client.Client, sandboxID string) {
// 	ctx := context.Background()
// 	cli.ContainerStop(ctx, sandboxID, nil)
// 	cli.ContainerRemove(ctx, sandboxID, types.ContainerRemoveOptions{})
// }

// // startSandbox creates a container for needed language and starts it.
// func startSandbox(ctx context.Context, cli *client.Client, language string) (string, error) {
// 	image := "code-runner/" + language

// 	resp, err := cli.ContainerCreate(ctx, &container.Config{
// 		Image:           image,
// 		OpenStdin:       true,
// 		NetworkDisabled: true,
// 	}, nil, nil, nil, "")
// 	if err != nil {
//     return "", fmt.Errorf("sandbox.startSandbox: failed to create a container: %w", err)
// 	}

// 	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
//     return "", fmt.Errorf("sandbox.startSandbox: failed to start the sandbox: %w", err)
// 	}

// 	return resp.ID, nil
// }

// // writeRunRequest writes payload in json format to contianer's stdin.
// func writeRunRequest(attach types.HijackedResponse, payload Request) error {
// 	payloadJson, err := json.Marshal(payload)
// 	}

// 	if _, err := attach.Conn.Write(append(payloadJson, '\n')); err != nil {
// 		return err
// 	}

// 	return nil
// }

// // readStd reads container's standard output and standard error.
// func readStd(attach types.HijackedResponse) ([]byte, []byte, error) {
// 	var stdout, stderr bytes.Buffer

// 	if _, err := stdcopy.StdCopy(&stdout, &stderr, attach.Reader); err != nil {
//     return nil, nil, fmt.Errorf("sandbox.readStd: failed to copy container's std: %w", err) 
// 	}

//   fmt.Println(stdout.String(), stderr.String())
// 	return stdout.Bytes(), stderr.Bytes(), nil
// }
