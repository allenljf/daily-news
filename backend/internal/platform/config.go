// Package platform provides process-level configuration and infrastructure setup.
package platform

import (
	"fmt"
	"net"
	"strconv"
)

// APIAddress converts Cloud Run's PORT value into the API listen address.
func APIAddress(port string) (string, error) {
	if port == "" {
		port = "8080"
	}

	parsed, err := strconv.ParseUint(port, 10, 16)
	if err != nil || parsed == 0 {
		return "", fmt.Errorf("invalid PORT %q", port)
	}

	return net.JoinHostPort("0.0.0.0", port), nil
}
