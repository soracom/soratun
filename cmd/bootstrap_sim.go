package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/soracom/soratun"
	"github.com/spf13/cobra"
)

var (
	provisioningAPIEndpointURL string
	requestParameters          string
	keysAPIEndpointURL         string
	signatureAlgorithm         string
	uiccInterfaceType          string
	portName                   string
	baudRate                   uint
	dataBits                   uint
	stopBits                   uint
	parityMode                 uint
	disableKeyCache            bool
	clearKeyCache              bool
	kryptonCliPath             string
)

func bootstrapSimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sim",
		Short: "Create virtual SIM which is associated with current SIM with SORACOM Krypton SIM authentication",
		Long:  "This command will create a new virtual SIM which is associated with current physical SIM, then create configuration for soratun. You need working \"krypton-cli\". See https://github.com/soracom/krypton-client-go for how to install.",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := bootstrap(&soratun.SimBootstrapper{
				KryptonCliPath: kryptonCliPath,
				Arguments:      buildKryptonCliArguments(),
			})
			if err != nil {
				log.Fatalf("failed to bootstrap: %v", err)
			}
		},
	}

	cmd.Flags().StringVar(&provisioningAPIEndpointURL, "provisioning-api-endpoint-url", "", "Use the specified URL as a Provisioning API endpoint.")
	cmd.Flags().StringVar(&requestParameters, "params", "", "Pass additional JSON parameters to the service request")
	cmd.Flags().StringVar(&keysAPIEndpointURL, "keys-api-endpoint-url", "", "Use the specified URL as a Keys API endpoint")
	cmd.Flags().StringVar(&signatureAlgorithm, "signature-algorithm", "SHA-256", "Algorithm for generating signature.")
	cmd.Flags().StringVar(&uiccInterfaceType, "interface", "autoDetect", "UICC Interface to use. Valid values are iso7816, comm, mmcli or autoDetect")
	cmd.Flags().StringVar(&portName, "port-name", "", "Port name of communication device (e.g. COM1 or /dev/tty1)")
	cmd.Flags().UintVar(&baudRate, "baud-rate", 57600, "Baud rate for communication device")
	cmd.Flags().UintVar(&dataBits, "data-bits", 8, "Data bits for communication device")
	cmd.Flags().UintVar(&stopBits, "stop-bits", 1, "Stop bits for communication device")
	cmd.Flags().UintVar(&parityMode, "parity-mode", 0, "Parity mode for communiation device. 0: None (default), 1: Odd, 2: Even")
	cmd.Flags().BoolVar(&disableKeyCache, "disable-key-cache", false, "Do not store authentication result to the key cache")
	cmd.Flags().BoolVar(&clearKeyCache, "clear-key-cache", false, "Remove all items in the key cache")
	cmd.Flags().StringVar(&kryptonCliPath, "krypton-cli-path", "/usr/local/bin/krypton-cli", "Path to krypton-cli")

	return cmd
}

func buildKryptonCliArguments() []string {
	var args []string
	args = append(args, []string{"-operation", "bootstrapArc"}...)
	args = append(args, []string{"-signature-algorithm", signatureAlgorithm}...)
	args = append(args, []string{"-interface", uiccInterfaceType}...)
	args = append(args, []string{"-baud-rate", fmt.Sprint(baudRate)}...)
	args = append(args, []string{"-data-bits", fmt.Sprint(dataBits)}...)
	args = append(args, []string{"-stop-bits", fmt.Sprint(stopBits)}...)
	args = append(args, []string{"-parity-mode", fmt.Sprint(parityMode)}...)
	if provisioningAPIEndpointURL != "" {
		args = append(args, []string{"-provisioning-api-endpoint-url", provisioningAPIEndpointURL}...)
	}
	if requestParameters != "" {
		args = append(args, []string{"-params", requestParameters}...)
	}
	if keysAPIEndpointURL != "" {
		args = append(args, []string{"-keys-api-endpoint-url", keysAPIEndpointURL}...)
	}
	if portName != "" {
		args = append(args, []string{"-port-name", portName}...)
	}
	if disableKeyCache {
		args = append(args, []string{"-disable-key-cache"}...)
	}
	if clearKeyCache {
		args = append(args, []string{"-clear-key-cache"}...)
	}
	if v := os.Getenv("SORACOM_VERBOSE"); v != "" {
		args = append(args, []string{"-debug"}...)
	}
	return args
}
