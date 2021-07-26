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
					ArcAllowedIPs: []*soratun.IPNet{{
						IP:   net.IPv4(100, 127, 0, 0),
						Mask: net.IPv4Mask(255, 255, 0, 0),
					}},
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
