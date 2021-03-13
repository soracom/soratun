package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/soracom/soratun"
	"github.com/spf13/cobra"
)

var (
	// Config holds SORACOM Arc client configuration.
	Config *soratun.Config
	// configPath holds path to SORACOM Arc client configuration file.
	configPath string
	// ctx is a context object for internal use to prove (default: Background()).
	ctx = context.Background()
)

// RootCmd defines soratun, a top level command.
var RootCmd = &cobra.Command{
	Use:   "soratun [command]",
	Short: "soratun -- SORACOM Arc Client",
}

func init() {
	RootCmd.PersistentFlags().StringVar(&configPath, "config", "arc.json", "Specify path to SORACOM Arc client configuration file")

	RootCmd.AddCommand(bootstrapCmd())
	RootCmd.AddCommand(configCmd())
	RootCmd.AddCommand(statusCmd())
	RootCmd.AddCommand(upCmd())
	RootCmd.AddCommand(versionCmd())
}

func initSoratun(_ *cobra.Command, _ []string) {
	config, err := readConfig(configPath)
	if err != nil {
		log.Fatalf("Error: %s\n", err)
	}
	Config = config

	if os.Getenv("__SORACOM_NO_DYNAMIC_CLIENT_SETUP_FOR_TEST") != "" {
		// NOTE:
		// This is for WireGuard integration testing purpose. It would inject the mocked client statically.
		fmt.Println("@@@@ DEVELOPMENT MODE @@@@ => dynamic client setup is suppressed for testing purpose")
		return
	}
}

func readConfig(path string) (*soratun.Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("faile to open config file: %s", path)
	}

	var config soratun.Config
	err = json.Unmarshal(b, &config)
	if err != nil {
		return nil, fmt.Errorf("error while reading config file: %s", err)
	}

	return &config, nil
}

func dumpWireGuardConfig(session *soratun.ArcSession) {
	var ips []string
	for _, ip := range session.ArcAllowedIPs {
		ips = append(ips, (*net.IPNet)(ip).String())
	}

	fmt.Printf(`--- WireGuard configuration ----------------------
[Interface]
Address = %s/32
PrivateKey = %s

[Peer]
PublicKey = %s
AllowedIPs = %s
Endpoint = %s:%d
--- End of WireGuard configuration ---------------
`,
		session.ArcClientPeerIpAddress,
		Config.PrivateKey,
		session.ArcServerPeerPublicKey,
		strings.Join(ips, ", "),
		session.ArcServerEndpoint.IP,
		session.ArcServerEndpoint.Port)
}
