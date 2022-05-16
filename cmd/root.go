package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

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
	RootCmd.AddCommand(completionCmd())
	RootCmd.AddCommand(configCmd())
	RootCmd.AddCommand(dumpWireGuardConfigCmd())
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

	if Config.Mtu == 0 {
		Config.Mtu = soratun.DefaultMTU
	}

	if Config.PersistentKeepalive == 0 {
		Config.PersistentKeepalive = soratun.DefaultPersistentKeepaliveInterval
	}

	if len(Config.AdditionalAllowedIPs) > 0 {
		Config.ArcSession.ArcAllowedIPs = append(Config.ArcSession.ArcAllowedIPs, Config.AdditionalAllowedIPs...)
	}

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
		return nil, fmt.Errorf("failed to open config file: %s", path)
	}

	var config soratun.Config
	err = json.Unmarshal(b, &config)
	if err != nil {
		return nil, fmt.Errorf("error while reading config file: %s", err)
	}

	return &config, nil
}
