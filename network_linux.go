package soratun

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/device"
	"net"
)

// ConfigureInterface create a new network interface with given SORACOM Arc configuration. Then setup routing table for allowedIPs.
func ConfigureInterface(iname string, config *Config) error {
	logger := device.NewLogger(
		config.LogLevel,
		fmt.Sprintf("(%s) ", iname),
	)

	logger.Verbosef("assign IP address: %s", config.ArcSession.ArcClientPeerIpAddress)
	iface, err := netlink.LinkByName(iname)
	if err != nil {
		return err
	}

	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   config.ArcSession.ArcClientPeerIpAddress,
			Mask: []byte{0xff, 0xff, 0xff, 0xff},
		},
		Label: "",
		Flags: 0,
		Scope: 0,
		Peer:  nil,
	}

	if err := netlink.AddrAdd(iface, addr); err != nil {
		return err
	}

	logger.Verbosef("set link up: %s", iname)
	if err := netlink.LinkSetUp(iface); err != nil {
		return err
	}

	for _, allowedIP := range config.ArcSession.ArcAllowedIPs {
		prefix, _ := allowedIP.Mask.Size()
		logger.Verbosef("add route: %s/%d", allowedIP.IP, prefix)
		route := netlink.Route{
			LinkIndex: iface.Attrs().Index,
			Scope:     netlink.SCOPE_LINK,
			Dst:       (*net.IPNet)(allowedIP),
		}
		if err := netlink.RouteReplace(&route); err != nil {
			return err
		}
	}

	return nil
}
