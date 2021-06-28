package cmd

import (
	"fmt"
	"github.com/soracom/soratun"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func upCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "up",
		Aliases: []string{"u"},
		Short:   "Setup SORACOM Arc interface",
		PreRun:  initSoratun,
		Run: func(cmd *cobra.Command, args []string) {
			if Config.ArcSession == nil {
				log.Fatal("Failed to determine connection information. Please bootstrap or create a new session from the user console.")
			}

			if v := os.Getenv("SORACOM_VERBOSE"); v != "" {
				fmt.Println("--- WireGuard configuration ----------------------")
				dumpWireGuardConfig()
				fmt.Println("--- End of WireGuard configuration ---------------")
			}

			soratun.Up(ctx, Config)
		},
	}
}
