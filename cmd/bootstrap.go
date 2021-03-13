package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/soracom/soratun"
	"github.com/spf13/cobra"
	"log"
	"path/filepath"
)

func bootstrapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Create virtual SIM and configure soratun",
		Long: `"soratun bootstrap" provides a set of commands which will (1) create a virtual SIM, and (2) create configuration for soratun.

COMMAND  CREATES                 AUTH METHOD          REQUIREMENTS                                    CONNECTIVITY PLATFORM
-------- ----------------------- -------------------- ----------------------------------------------- ------------ -------------
authkey  Standalone virtual SIM  SORACOM API AuthKey  SORACOM API Auth Key                            Any          Linux, macOS
-------- ----------------------- -------------------- ----------------------------------------------- ------------ -------------
cellular Virtual SIM which is    cellular connection  Active SORACOM Air cellular connection          Cellular     Linux, macOS
-------- associated with         -------------------- ----------------------------------------------- ------------ -------------
sim      current SIM             SIM Authentication   Compatible modem/SIM card reader, and OS setup  Any          Linux
-------- ----------------------- -------------------- ----------------------------------------------- ------------ -------------
`,
	}

	cmd.AddCommand(bootstrapAuthKeyCmd())
	cmd.AddCommand(bootstrapCellularCmd())
	cmd.AddCommand(bootstrapSimCmd())

	return cmd
}

func bootstrap(bootstrapper soratun.Bootstrapper) (*soratun.Config, error) {
	// if current config exists, pass it to the bootstrapper to merge if applicable
	currentConfig, _ := readConfig(configPath)
	config, err := bootstrapper.Execute(currentConfig)
	if err != nil {
		return nil, err
	}

	b, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, err
	}

	err = writeConfigurationToFile(string(b))
	if err != nil {
		return nil, err
	}

	if config.SimId != "" {
		fmt.Printf("Virtual subscriber SIM ID: %s\n", config.SimId)
	}

	printConfigurationFilePath()

	return config, nil
}

func printConfigurationFilePath() {
	path, err := filepath.Abs(configPath)
	if err != nil {
		log.Fatalf("Failed to get path to configuration file: %v\n", err)
	}
	fmt.Printf("Created/updated configuration file: %s\n", path)
}
