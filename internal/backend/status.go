package backend

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const protonInterface = "proton0"

var ErrNmcliNotFound = errors.New("nmcli: binary not found in PATH")

type ConnectionStatus struct {
	Connected bool
	Server    string
	Location  string
	Load      int
	Protocol  string
}

func Status() (ConnectionStatus, error) {
	connected, err := isConnectedViaNetworkManager()
	if err != nil {
		return ConnectionStatus{}, err
	}
	if !connected {
		return ConnectionStatus{Connected: false}, nil
	}

	res, err := Run(context.Background(), "status")
	if err != nil {
		return ConnectionStatus{}, err
	}

	status := parseStatus(res.Stdout)
	status.Connected = true
	return status, nil
}

func IsConnected() (bool, error) {
	return isConnectedViaNetworkManager()
}

func isConnectedViaNetworkManager() (bool, error) {
	if _, err := exec.LookPath("nmcli"); err != nil {
		return false, ErrNmcliNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "nmcli", "-t", "-f", "DEVICE", "connection", "show", "--active")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("nmcli: %w (stderr: %q)", err, strings.TrimSpace(stderr.String()))
	}

	for device := range strings.SplitSeq(strings.TrimSpace(stdout.String()), "\n") {
		if device == protonInterface {
			return true, nil
		}
	}

	return false, nil
}

func parseStatus(output string) ConnectionStatus {
	var status ConnectionStatus

	for line := range strings.SplitSeq(output, "\n") {
		key, value, found := strings.Cut(strings.TrimSpace(line), ": ")
		if !found {
			continue
		}

		switch key {
		case "Status":
			status.Connected = value == "Connected"
		case "Server":
			if server, location, ok := strings.Cut(value, " in "); ok {
				status.Server = server
				status.Location = location
			} else {
				status.Server = value
			}
		case "Load":
			if load, err := strconv.Atoi(strings.TrimSuffix(value, "%")); err == nil {
				status.Load = load
			}
		case "Protocol":
			status.Protocol = value
		}
	}

	return status
}
