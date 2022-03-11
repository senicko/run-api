package sandbox

import (
	"bufio"
	"bytes"
	"context"
	"net"

	"github.com/docker/docker/pkg/stdcopy"
)

func writeStd(ctx context.Context, writer net.Conn, payload []byte) error {
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

func readStd(reader *bufio.Reader) ([]byte, []byte, error) {
	var stdout, stderr bytes.Buffer

	if _, err := stdcopy.StdCopy(&stdout, &stderr, reader); err != nil {
		return nil, nil, err
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}
