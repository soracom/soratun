package soratun

import (
	"fmt"
	"golang.zx2c4.com/wireguard/device"
	"os/exec"
	"strings"
)

const (
	// LogLevelVerbose is an alias for WireGuard device logger equivalent.
	LogLevelVerbose = device.LogLevelVerbose
	// LogLevelError is an alias for WireGuard device logger equivalent.
	LogLevelError = device.LogLevelError
	// LogLevelSilent is an alias for WireGuard device logger equivalent.
	LogLevelSilent = device.LogLevelSilent
	// DefaultPersistentKeepaliveInterval defines WireGuard persistent keepalive interval to SORACOM Arc.
	DefaultPersistentKeepaliveInterval = 60
	// DefaultMTU is MTU for the configured interface.
	DefaultMTU = device.DefaultMTU
)

func runCommand(c []string) (string, error) {
	result, err := exec.Command(c[0], c[1:]...).CombinedOutput()

	if err != nil {
		return "", fmt.Errorf(
			"error while running \"%s\" with %s, output: '%s'",
			strings.Join(c, " "),
			err,
			strings.TrimSpace(string(result)),
		)
	}

	return fmt.Sprintf("'%s'\n", strings.TrimSpace(string(result))), nil
}

func replaceInterfaceName(command []string, iname string) []string {
	var replaced []string
	for _, s := range command {
		replaced = append(replaced, strings.Replace(s, "%i", iname, -1))
	}
	return replaced
}
