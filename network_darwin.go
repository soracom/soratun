package soratun

import (
	"fmt"

	"golang.zx2c4.com/wireguard/device"
)

// ConfigureInterface create a new network interface with given SORACOM Arc configuration. Then setup routing table for allowedIPs.
func ConfigureInterface(iname string, config *Config) error {
	logger := device.NewLogger(
		config.LogLevel,
		fmt.Sprintf("(%s) ", iname),
	)

	command := []string{"sudo", "ifconfig", iname, config.ArcSession.ArcClientPeerIpAddress.String(), config.ArcSession.ArcClientPeerIpAddress.String()}
	logger.Verbosef("assign IP address: %s", command)
	_, err := runCommand(command)
	if err != nil {
		return err
	}

	for _, allowedIP := range config.ArcSession.ArcAllowedIPs {
		prefix, _ := allowedIP.Mask.Size()
		if prefix == 32 {
			command = []string{"sudo", "route", "add", "-host", allowedIP.IP.String(), "-interface", iname}
		} else {
			command = []string{"sudo", "route", "add", "-net", fmt.Sprintf("%s/%d", allowedIP.IP, prefix), "-interface", iname}
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
