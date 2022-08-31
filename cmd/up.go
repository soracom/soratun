package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/soracom/soratun"
	"github.com/spf13/cobra"
)

var (
	mtu                  int
	persistentKeepalive  int
	additionalAllowedIPs string
	readStdin            bool
)

func upCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "up",
		Aliases: []string{"u"},
		Short:   "Setup SORACOM Arc interface",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if readStdin {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					log.Fatalf("Failed to read configuration from stdin: %v", err)
				}

				var config soratun.Config
				err = json.Unmarshal(b, &config)
				if err != nil {
					log.Fatalf("Failed to read configuration from stdin: %v", err)
				}
				Config = &config
			} else {
				initSoratun(cmd, args)
			}

			// override only if the flag was explicitly set
			if cmd.Flags().Changed("mtu") {
				Config.Mtu = mtu
			}

			if cmd.Flags().Changed("persistent-keepalive") {
				Config.PersistentKeepalive = persistentKeepalive
			}

			if Config.ArcSession == nil {
				log.Fatal("Failed to determine connection information. Please bootstrap or create a new session from the user console.")
			}

			if additionalAllowedIPs != "" {
				for _, s := range strings.Split(additionalAllowedIPs, ",") {
					_, ipnet, err := net.ParseCIDR(strings.TrimSpace(s))
					if err != nil {
						log.Fatalf("Invalid CIDR is set for \"--additional-allowd-ips\": %v", err)
					}
					Config.ArcSession.ArcAllowedIPs = append(Config.ArcSession.ArcAllowedIPs, &soratun.IPNet{
						IP:   ipnet.IP,
						Mask: ipnet.Mask,
					})
				}
			}

			if v := os.Getenv("SORACOM_VERBOSE"); v != "" {
				fmt.Fprintln(os.Stderr, "--- WireGuard configuration ----------------------")
				dumpWireGuardConfig(true, os.Stderr)
				fmt.Fprintln(os.Stderr, "--- End of WireGuard configuration ---------------")
			}

			soratun.Up(ctx, Config)
		},
	}

	cmd.Flags().IntVar(&mtu, "mtu", soratun.DefaultMTU, "MTU for the interface, which will override arc.json#mtu value")
	cmd.Flags().IntVar(&persistentKeepalive, "persistent-keepalive", soratun.DefaultPersistentKeepaliveInterval, "WireGuard \"PersistentKeepalive\" for the SORACOM Arc server, which will override arc.json#persistentKeepalive value")
	cmd.Flags().StringVar(&additionalAllowedIPs, "additional-allowed-ips", "", "Comma separated string of additional WireGuard allowed CIDRs, which will be added to arc.json#additionalAllowedIPs array")
	cmd.Flags().BoolVar(&readStdin, "read-stdin", false, "read configuration from stdin, ignoring --config setting")

	return cmd
}
