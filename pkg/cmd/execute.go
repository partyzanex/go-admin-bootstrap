package cmd

import (
	"bytes"
	"context"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func Execute(ctx context.Context, workDir, binary string, args ...string) (*bytes.Buffer, error) {
	stdoutBuf, stderrBuf := new(bytes.Buffer), new(bytes.Buffer)

	command := exec.CommandContext(ctx, binary, args...)
	command.Dir = workDir
	command.Stdout = stdoutBuf
	command.Stderr = stderrBuf

	if err := command.Start(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			err = ErrBinNotFound
		}

		return nil, errors.Wrapf(err, "cannot start command %q with args %s", binary, joinArgs(args))
	}

	errs := make(chan error, 1)

	go func() {
		errs <- command.Wait()
	}()

	select {
	case <-ctx.Done():
		err := command.Process.Kill()
		if err != nil {
			return combineBuffers(stdoutBuf, stderrBuf), errors.Wrap(err, "cannot kill process")
		}

		return combineBuffers(stdoutBuf, stderrBuf), ErrTimeOut
	case err := <-errs:
		if err != nil {
			return combineBuffers(stdoutBuf, stderrBuf), errors.Wrapf(err, "cmd.Wait: command %q, args %q",
				binary, joinArgs(args),
			)
		}
	}

	return stdoutBuf, nil
}

func combineBuffers(stdoutBuf, stderrBuf *bytes.Buffer) *bytes.Buffer {
	output := new(bytes.Buffer)

	if stdoutBuf.Len() > 0 {
		output.WriteString("stdout:\n")
		output.Write(stdoutBuf.Bytes())
		output.WriteString("\n")
	}

	if stderrBuf.Len() > 0 {
		output.WriteString("stderr:\n")
		output.Write(stderrBuf.Bytes())
		output.WriteString("\n")
	}

	return output
}

func joinArgs(args []string) string {
	return strings.Join(args, ", ")
}
