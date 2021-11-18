package soratun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
)

// SimBootstrapper defines bootstrap method with SORACOM Krypton SIM authentication. Needs krypton-cli installed.
type SimBootstrapper struct {
	KryptonCliPath string
	Arguments      []string
}

// Execute calls SORACOM Krypton CLI to create a new virtual subscriber which is associated with current physical SIM.
func (b *SimBootstrapper) Execute(config *Config) (*Config, error) {
	if _, err := os.Stat(b.KryptonCliPath); os.IsNotExist(err) {
		return nil, err
	}

	if v := os.Getenv("SORACOM_VERBOSE"); v != "" {
		fmt.Fprintf(os.Stderr, "Running %s %s\n", b.KryptonCliPath, b.Arguments)
	}

	// if no config, create a blank, then replace keys and ArcSession with new
	if config == nil {
		config = &Config{
			PrivateKey:           Key{},
			PublicKey:            Key{},
			SimId:                "",
			LogLevel:             LogLevelVerbose,
			EnableMetrics:        true,
			Interface:            DefaultInterfaceName(),
			AdditionalAllowedIPs: nil,
			Profile:              nil,
			ArcSession:           nil,
		}
	}

	cmd := exec.Command(b.KryptonCliPath, b.Arguments...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error while running %s: %s\n%s", b.KryptonCliPath, err, &stderr)
	}

	var arcSession ArcSession
	err = json.Unmarshal(stdout.Bytes(), &arcSession)
	if err != nil {
		return nil, fmt.Errorf("error while reading response from krypton-cli: %s\nkrypton-cli output:\n-----\n%s", err, stdout.String())
	}

	// Since soratun.ArcSession marshaler omits ArcClientPeerPrivateKey, we can simply and safely marshal response from
	// Krypton CLI, and print it to stdout.
	t, err := json.Marshal(&arcSession)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling response from krypton-cli: %s", err)
	}

	if v := os.Getenv("SORACOM_VERBOSE"); v != "" {
		fmt.Fprintf(os.Stderr, "Got response from %s: %s\n", b.KryptonCliPath, t)
	}

	config.PrivateKey = arcSession.ArcClientPeerPrivateKey
	config.PublicKey = (Key)(arcSession.ArcClientPeerPrivateKey.AsWgKey().PublicKey())
	config.ArcSession = &arcSession
	return config, nil
}
