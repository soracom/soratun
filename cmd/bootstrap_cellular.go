package cmd

import (
	"log"

	"github.com/soracom/soratun"
	"github.com/spf13/cobra"
)

var kryptonCellularEndpoint string

func bootstrapCellularCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cellular",
		Short: "Create virtual SIM which is associated with current SIM with active cellular connection",
		Long:  "This command will create a new virtual SIM which is associated with current physical SIM, then create configuration for soratun. Need active SORACOM Air for Cellular connection.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := bootstrap(&soratun.CellularBootstrapper{Endpoint: kryptonCellularEndpoint})
			if err != nil {
				log.Fatalf("failed to bootstrap: %v", err)
			}
		},
	}

	cmd.Flags().StringVar(&kryptonCellularEndpoint, "endpoint", "https://krypton.soracom.io:8036", "Specify SORACOM Krypton Provisioning API endpoint.")
	cmd.Flags().BoolVar(&dumpConfig, "dump-config", false, "dump configuration to stdout, ignoring --config setting")

	return cmd
}
