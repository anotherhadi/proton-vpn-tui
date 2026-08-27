package backend

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const DefaultTimeout = 15 * time.Second

var BinaryName = "protonvpn"

var ErrTimeout = errors.New("protonvpn: command timeout")

var ErrNotFound = errors.New("protonvpn: binary not found in PATH")

type Result struct {
	Stdout string
	Stderr string
}

type CommandError struct {
	Result *Result
	Err    error
}

const unexpectedErrorBanner = "An unexpected error occurred. Please try again."

func (e *CommandError) Error() string {
	if strings.HasPrefix(e.Result.Stderr, unexpectedErrorBanner) ||
		strings.HasPrefix(e.Result.Stdout, unexpectedErrorBanner) {
		return unexpectedErrorBanner
	}
	if e.Result.Stderr != "" {
		return fmt.Sprintf("%v (stderr: %q)", e.Err, e.Result.Stderr)
	}
	if e.Result.Stdout != "" {
		return e.Result.Stdout
	}
	return e.Err.Error()
}

func (e *CommandError) Unwrap() error {
	return e.Err
}

func IsAuthRequired(err error) bool {
	var cmdErr *CommandError
	if !errors.As(err, &cmdErr) {
		return false
	}
	return strings.Contains(cmdErr.Result.Stderr, "protonvpn signin") ||
		strings.Contains(cmdErr.Result.Stdout, "protonvpn signin")
}

func IsFreePlanRestricted(err error) bool {
	var cmdErr *CommandError
	if !errors.As(err, &cmdErr) {
		return false
	}
	return strings.Contains(cmdErr.Result.Stderr, "not available on the free plan") ||
		strings.Contains(cmdErr.Result.Stdout, "not available on the free plan")
}

func Available() error {
	if _, err := exec.LookPath(BinaryName); err != nil {
		return ErrNotFound
	}
	return nil
}

func Run(ctx context.Context, args ...string) (*Result, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), DefaultTimeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, BinaryName, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	res := &Result{
		Stdout: strings.TrimSpace(stdout.String()),
		Stderr: strings.TrimSpace(stderr.String()),
	}

	if ctx.Err() == context.DeadlineExceeded {
		return res, &CommandError{Result: res, Err: ErrTimeout}
	}

	if runErr != nil {
		return res, &CommandError{Result: res, Err: runErr}
	}

	if strings.HasPrefix(res.Stdout, "Error:") {
		return res, &CommandError{Result: res, Err: errors.New(res.Stdout)}
	}

	return res, nil
}
