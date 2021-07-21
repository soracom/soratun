package soratun

import (
	"fmt"
	"strings"

	"golang.zx2c4.com/wireguard/device"
)

// ConfigureInterface create a new network interface with given SORACOM Arc configuration. Then setup routing table for allowedIPs.
func ConfigureInterface(iname string, config *Config) error {
	logger := device.NewLogger(
		config.LogLevel,
		fmt.Sprintf("(%s) ", iname),
	)

	command := strings.Split(fmt.Sprintf("sudo ifconfig %s %s %s", iname, config.ArcSession.ArcClientPeerIpAddress, config.ArcSession.ArcClientPeerIpAddress), " ")
	logger.Verbosef("assign IP address: %s", command)
	_, err := runCommand(command)
	if err != nil {
		return err
	}

	for _, allowedIP := range config.ArcSession.ArcAllowedIPs {
		prefix, _ := allowedIP.Mask.Size()
		if prefix == 32 {
			command = strings.Split(fmt.Sprintf("sudo route add -host %s -interface %s", allowedIP.IP, iname), " ")
		} else {
			command = strings.Split(fmt.Sprintf("sudo route add -net %s/%d -interface %s", allowedIP.IP, prefix, iname), " ")
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
