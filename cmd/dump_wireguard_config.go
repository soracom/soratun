package cmd

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"
)

func dumpWireGuardConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "wg-config",
		Short:  "Dump soratun configuration file as WireGuard format",
		Args:   cobra.NoArgs,
		PreRun: initSoratun,
		Run: func(cmd *cobra.Command, args []string) {
			dumpWireGuardConfig(false)
		},
	}
}

func dumpWireGuardConfig(mask bool) {
	var ips []string
	for _, ip := range Config.ArcSession.ArcAllowedIPs {
		ips = append(ips, (*net.IPNet)(ip).String())
	}

	privateKey := (Config.PrivateKey).String()
	if mask {
		privateKey = "(hidden)"
	}

	postUp := ""
	postDown := ""
	if len(Config.PostUp) > 0 {
		for _, com := range Config.PostUp {
			if len(com) == 0 || com[0] == "" {
				continue
			}
			postUp = fmt.Sprintf("%sPostUp = %s\n", postUp, strings.Join(com, " "))
		}
	}
	if len(Config.PostDown) > 0 {
		for _, com := range Config.PostDown {
			if len(com) == 0 || com[0] == "" {
				continue
			}
			postDown = fmt.Sprintf("%sPostDown = %s\n", postDown, strings.Join(com, " "))
		}
	}

	fmt.Printf(`[Interface]
Address = %s/32
PrivateKey = %s
MTU = %d
%s%s
[Peer]
PublicKey = %s
AllowedIPs = %s
Endpoint = %s:%d
PersistentKeepalive = %d
`,
		Config.ArcSession.ArcClientPeerIpAddress,
		privateKey,
		Config.Mtu,
		postUp,
		postDown,
		Config.ArcSession.ArcServerPeerPublicKey,
		strings.Join(ips, ", "),
		Config.ArcSession.ArcServerEndpoint.IP,
		Config.ArcSession.ArcServerEndpoint.Port,
		Config.PersistentKeepalive,
	)
}
