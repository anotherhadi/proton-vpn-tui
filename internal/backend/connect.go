package backend

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

func connectArgs(country, city string, p2p, secureCore, tor bool) []string {
	args := []string{"connect"}
	if country != "" {
		args = append(args, "--country", country)
	}
	if city != "" {
		args = append(args, "--city", city)
	}
	if p2p {
		args = append(args, "--p2p")
	}
	if secureCore {
		args = append(args, "--securecore")
	}
	if tor {
		args = append(args, "--tor")
	}
	return args
}

func ConnectFastest(country, city string, p2p, secureCore, tor bool) (err error) {
	_, err = Run(context.Background(), connectArgs(country, city, p2p, secureCore, tor)...)
	return
}

func ConnectRandom(country, city string, p2p, secureCore, tor bool) (err error) {
	args := append(connectArgs(country, city, p2p, secureCore, tor), "--random")
	_, err = Run(context.Background(), args...)
	return
}

func ConnectId(id string) (err error) {
	if !isValidServerId(id) {
		return fmt.Errorf("protonvpn: invalid server id %q", id)
	}
	_, err = Run(context.Background(), "connect", id)
	return
}

func ConnectAny() (err error) {
	_, err = Run(context.Background(), "connect")
	return
}

func Disconnect() (err error) {
	res, err := Run(context.Background(), "disconnect")
	if err != nil && res != nil && strings.HasPrefix(res.Stdout, "Disconnected") {
		// The official CLI sometimes reports a successful disconnect and
		// then, in a separate step (e.g. cleanup/telemetry), fails and
		// exits non-zero with the generic error banner. The VPN is
		// already disconnected at that point, so treat it as success.
		return nil
	}
	return err
}

var serverIdPattern = regexp.MustCompile(`^[A-Z]{2}(-[A-Z0-9]{2,4})?#[0-9]+(-TOR)?$`)

func isValidServerId(id string) bool {
	return serverIdPattern.MatchString(id)
}

func IsServerID(s string) bool {
	return isValidServerId(s)
}
