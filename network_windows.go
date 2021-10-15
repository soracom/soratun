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

	command := []string{"netsh", "interface", "ip", "set", "address", iname, "static", config.ArcSession.ArcClientPeerIpAddress.String(), "255.255.255.255"}
	logger.Verbosef("assign IP address: %s", command)
	_, err := runCommand(command)
	if err != nil {
		return err
	}

	for _, allowedIP := range config.ArcSession.ArcAllowedIPs {
		m := allowedIP.Mask
		mask := fmt.Sprintf("%d.%d.%d.%d", m[0], m[1], m[2], m[3])
		command = []string{"netsh", "routing", "ip", "add", "persistentroute", "dest="+allowedIP.IP.String(), "mask="+mask, "name="+iname}
		logger.Verbosef("update routing table: %s", command)
		result, err := runCommand(command)
		if err != nil {
			return err
		}
		logger.Verbosef("%s", result)
	}
	return nil
}
