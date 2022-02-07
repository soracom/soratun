package soratun

import (
	"os"
)

// CellularBootstrapper defines bootstrap method with SORACOM Krypton cellular authentication. Needs active cellular connection.
type CellularBootstrapper struct {
	Endpoint string
}

// Execute calls SORACOM Krypton Provisioning API cellular endpoint to create a new virtual subscriber which is associated with current physical SIM.
func (b *CellularBootstrapper) Execute(config *Config) (*Config, error) {
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
			Mtu:                  DefaultMTU,
			PersistentKeepalive:  DefaultPersistentKeepaliveInterval,
			Profile:              nil,
			ArcSession:           nil,
		}
	}
	client := NewDefaultSoracomKryptonClient(&KryptonClientConfig{Endpoint: b.Endpoint})

	if v := os.Getenv("SORACOM_VERBOSE"); v != "" {
		client.SetVerbose(true)
	}

	arcSession, err := client.Bootstrap()
	if err != nil {
		return nil, err
	}
	config.PrivateKey = arcSession.ArcClientPeerPrivateKey
	config.PublicKey = (Key)(arcSession.ArcClientPeerPrivateKey.AsWgKey().PublicKey())
	config.ArcSession = arcSession

	return config, nil
}
