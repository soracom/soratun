package soratun

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"golang.zx2c4.com/wireguard/device"
)

// ConfigureInterface create a new network interface with given SORACOM Arc configuration. Then setup routing table for allowedIPs.
func ConfigureInterface(iname string, config *Config) error {
	logger := device.NewLogger(
		config.LogLevel,
		fmt.Sprintf("(%s) ", iname),
	)

	command := fmt.Sprintf("sudo ifconfig %s %s %s", iname, config.ArcSession.ArcClientPeerIpAddress, config.ArcSession.ArcClientPeerIpAddress)
	logger.Verbosef("assign IP address: %s", command)
	_, err := runCommand(command)
	if err != nil {
		return err
	}

	for _, allowedIP := range config.ArcSession.ArcAllowedIPs {
		prefix, _ := allowedIP.Mask.Size()
		if prefix == 32 {
			command = fmt.Sprintf("sudo route add -host %s -interface %s", allowedIP.IP, iname)
		} else {
			command = fmt.Sprintf("sudo route add -net %s/%d -interface %s", allowedIP.IP, prefix, iname)
		}
		logger.Verbosef("update routing table: %s", command)
		result, err := runCommand(command)
		if err != nil {
			return err
		}
		logger.Verbosef("%s", result)
	}
	return nil
}

func runCommand(s string) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", s)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("error while setting up \"%s\"", s)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("error while starting \"%s\" %s", s, err)
	}

	result, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", fmt.Errorf("error while reading output from \"%s\"", s)
	}

	return fmt.Sprintf("'%s'\n", strings.TrimSpace(string(result))), nil
}
