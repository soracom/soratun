package cmd

import (
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/soracom/soratun"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var (
	authKeyId string
	authKey   string
	coverage  string
	endpoints = []string{"https://g.api.soracom.io", "https://api.soracom.io"}
)

func bootstrapAuthKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authkey",
		Short: "Create standalone virtual SIM with SORACOM API AuthKey",
		Long:  "This command will create a new virtual SIM which is not associated with any physical SIM, then create configuration for soratun. If configuration file (arc.json by default, or specified by --config flag) contains \"profile\", that information will be used. Else if insufficient or no flags is provided, the command will guide your setup through interactive wizard.",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var profile *soratun.Profile

			// reuse current profile information
			currentConfig, err := readConfig(configPath)
			if err == nil && currentConfig != nil && currentConfig.Profile != nil {
				profile = currentConfig.Profile
			}

			if profile == nil {
				if cmd.Flag("auth-key-id").Value.String() == "" ||
					cmd.Flag("auth-key").Value.String() == "" ||
					cmd.Flag("coverage-type").Value.String() == "" {
					fmt.Println("Not enough information to bootstrap. Launching wizard.")
					profile, err = collectProfileInformationInteractive()
					if err != nil {
						log.Fatalf("Error while setup: %v\n", err)
					}
				} else {
					profile, err = collectProfileInformationFromFlags()
					if err != nil {
						log.Fatalf("Error while setup: %v\n", err)
					}
				}
			}

			_, err = bootstrap(&soratun.AuthKeyBootstrapper{
				Profile: profile,
			})
			if err != nil {
				log.Fatalf("failed to bootstrap: %v", err)
			}
		},
	}

	cmd.Flags().StringVar(&authKeyId, "auth-key-id", "", "SORACOM API auth key ID")
	cmd.Flags().StringVar(&authKey, "auth-key", "", "SORACOM API auth key")
	cmd.Flags().StringVar(&coverage, "coverage-type", "", "Specify coverage type, 'g' for Global, 'jp' for Japan")

	return cmd
}

func collectProfileInformationInteractive() (*soratun.Profile, error) {
	authKeyId, err := askInput(promptui.Prompt{
		Label: "SORACOM API auth key ID (starts with \"keyId-\")",
		Validate: func(input string) error {
			if !strings.HasPrefix(input, "keyId-") {
				return errors.New("auth key ID should start with \"keyId-\"")
			}
			return nil
		},
		Mask: '*',
	})
	if err != nil {
		return nil, err
	}

	authKey, err = askInput(promptui.Prompt{
		Label: "SORACOM API auth key (starts with \"secret-\")",
		Validate: func(input string) error {
			if !strings.HasPrefix(input, "secret-") {
				return errors.New("auth key should start with \"secret-\"")
			}
			return nil
		},
		Mask: '*',
	})
	if err != nil {
		return nil, err
	}

	selected, err := askOne(promptui.Select{
		Label: "Coverage to create a new virtual SIM",
		Items: []string{
			"Global coverage (g.api.soracom.io)",
			"Japan coverage (api.soracom.io)",
		},
	})
	if err != nil {
		return nil, err
	}

	return &soratun.Profile{
		AuthKey:   authKey,
		AuthKeyID: authKeyId,
		Endpoint:  endpoints[selected],
	}, nil
}

func collectProfileInformationFromFlags() (*soratun.Profile, error) {
	if !strings.HasPrefix(authKeyId, "keyId-") {
		return nil, errors.New("auth key ID should start with \"keyId-\"")
	}

	if !strings.HasPrefix(authKey, "secret-") {
		return nil, errors.New("auth key should start with \"secret-\"")
	}

	endpoint := endpoints[1]
	if strings.HasPrefix(coverage, "g") {
		endpoint = endpoints[0]
	}

	return &soratun.Profile{
		AuthKey:   authKey,
		AuthKeyID: authKeyId,
		Endpoint:  endpoint,
	}, nil
}

func askInput(prompt promptui.Prompt) (string, error) {
	res, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("invalid input: %s", err)
	}
	return res, nil
}

func askOne(prompt promptui.Select) (int, error) {
	selected, _, err := prompt.Run()
	if err != nil {
		return -1, fmt.Errorf("invalid selection: %s", err)
	}
	return selected, nil
}
