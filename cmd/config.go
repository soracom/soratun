package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/soracom/soratun"
	"github.com/spf13/cobra"
)

func configCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Create initial soratun configuration file without bootstrapping",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			localhost := net.IPv4(127, 0, 0, 1)

			config := &soratun.Config{
				PrivateKey:           soratun.Key{},
				PublicKey:            soratun.Key{},
				SimId:                "890000xxxxxxxxxxxxx",
				LogLevel:             soratun.LogLevelVerbose,
				EnableMetrics:        true,
				Interface:            soratun.DefaultInterfaceName(),
				AdditionalAllowedIPs: nil,
				ArcSession: &soratun.ArcSession{
					ArcServerPeerPublicKey: soratun.Key{},
					ArcServerEndpoint: &soratun.UDPAddr{
						IP:          localhost,
						Port:        11010,
						RawEndpoint: []byte("localhost:11010"),
					},
					/*
					 * The following `ArcAllowedIPs` represents the following `AllowedIPs` configuration:
					 *
					 * ```
					 * AllowedIPs = 100.127.0.0/21, 100.127.8.0/23, 100.127.10.0/28, 100.127.10.17/32, 100.127.10.18/31, 100.127.10.20/30, 100.127.10.24/29, 100.127.10.32/27, 100.127.10.64/26, 100.127.10.128/25, 100.127.11.0/24, 100.127.12.0/22, 100.127.16.0/20, 100.127.32.0/19, 100.127.64.0/18, 100.127.128.0/17
					 * ```
					 *
					 * This means `AllowedIPs = +100.127.0.0/16, -100.127.10.16/32`, because `100.127.10.16/32` is the Soracom Napter's source IP when it uses Napter over Soracom Air,
					 * so Soracom Arc has to ignore this IP address in order to keep the symmetry network for Soracom Napter over Soracom Air.
					 * Just FYI, if it uses Soracom Napter over Soracom Arc, instead of Soracom Air, then the Soracom Napter's source is `100.127.10.17/32`
					 */
					ArcAllowedIPs: []*soratun.IPNet{
						{
							// 100.127.0.0/21
							IP:   net.IPv4(100, 127, 0, 0),
							Mask: net.IPv4Mask(255, 255, 248, 0),
						},
						{
							// 100.127.8.0/23
							IP:   net.IPv4(100, 127, 8, 0),
							Mask: net.IPv4Mask(255, 255, 254, 0),
						},
						{
							// 100.127.10.0/28
							IP:   net.IPv4(100, 127, 10, 0),
							Mask: net.IPv4Mask(255, 255, 255, 240),
						},
						{
							// 100.127.10.17/32
							IP:   net.IPv4(100, 127, 10, 17),
							Mask: net.IPv4Mask(255, 255, 255, 255),
						},
						{
							// 100.127.10.18/31
							IP:   net.IPv4(100, 127, 10, 18),
							Mask: net.IPv4Mask(255, 255, 255, 254),
						},
						{
							// 100.127.10.20/30
							IP:   net.IPv4(100, 127, 10, 20),
							Mask: net.IPv4Mask(255, 255, 255, 252),
						},
						{
							// 100.127.10.24/29
							IP:   net.IPv4(100, 127, 10, 24),
							Mask: net.IPv4Mask(255, 255, 255, 248),
						},
						{
							// 100.127.10.32/27
							IP:   net.IPv4(100, 127, 10, 32),
							Mask: net.IPv4Mask(255, 255, 255, 224),
						},
						{
							// 100.127.10.64/26
							IP:   net.IPv4(100, 127, 10, 64),
							Mask: net.IPv4Mask(255, 255, 255, 192),
						},
						{
							// 100.127.10.128/25
							IP:   net.IPv4(100, 127, 10, 128),
							Mask: net.IPv4Mask(255, 255, 255, 128),
						},
						{
							// 100.127.11.0/24
							IP:   net.IPv4(100, 127, 11, 0),
							Mask: net.IPv4Mask(255, 255, 255, 0),
						},
						{
							// 100.127.12.0/22
							IP:   net.IPv4(100, 127, 12, 0),
							Mask: net.IPv4Mask(255, 255, 252, 0),
						},
						{
							// 100.127.16.0/20
							IP:   net.IPv4(100, 127, 16, 0),
							Mask: net.IPv4Mask(255, 255, 240, 0),
						},
						{
							// 100.127.32.0/19
							IP:   net.IPv4(100, 127, 32, 0),
							Mask: net.IPv4Mask(255, 255, 224, 0),
						},
						{
							// 100.127.64.0/18
							IP:   net.IPv4(100, 127, 64, 0),
							Mask: net.IPv4Mask(255, 255, 192, 0),
						},
						{
							// 100.127.128.0/17
							IP:   net.IPv4(100, 127, 128, 0),
							Mask: net.IPv4Mask(255, 255, 128, 0),
						},
					},
					ArcClientPeerIpAddress: localhost,
				},
			}

			b, err := json.MarshalIndent(config, "", "  ")
			if err != nil {
				log.Fatalf("failed to convert config to JSON: %v", err)
			}

			fmt.Println(string(b))
		},
	}
}

func writeConfigurationToFile(conf string) error {
	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Fatal("failed to close file", err)
		}
	}()

	err = os.Chmod(f.Name(), 0600)
	if err != nil {
		return err
	}

	_, err = f.WriteString(conf + "\n")
	if err != nil {
		return err
	}

	return nil
}
