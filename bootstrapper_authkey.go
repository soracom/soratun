package soratun

import (
	"fmt"
	"os"
)

// AuthKeyBootstrapper defines bootstrap method with SORACOM API authentication. Needs Profile information.
type AuthKeyBootstrapper struct {
	Profile *Profile
}

// Execute calls SORACOM API to create a new standalone virtual subscriber.
func (b *AuthKeyBootstrapper) Execute(config *Config) (*Config, error) {
	client, err := NewDefaultSoracomClient(*b.Profile)
	if err != nil {
		return nil, err
	}

	if v := os.Getenv("SORACOM_VERBOSE"); v != "" {
		client.SetVerbose(true)
	}

	if config == nil {
		// if no config, bootstrap with API call
		sim, err := client.CreateVirtualSim()
		if err != nil {
			return nil, err
		}

		privateKey, err := NewKey(sim.Profiles[sim.SimId].ArcClientPeerPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("virtual SIM/subscriber %s was created but failed to create a configuration. "+
				"due to unexpected ArcClientPeerPrivateKey received. "+
				"Please open SORACOM User Console at https://console.soracom.io and check virtual SIM status. "+
				"You have to create arc.json manually", sim.SimId)
		}

		publicKey, err := NewKey(sim.Profiles[sim.SimId].ArcClientPeerPublicKey)
		if err != nil {
			return nil, fmt.Errorf("virtual SIM/subscriber %s was created but failed to create a configuration. "+
				"due to unexpected ArcClientPeerPublicKey received. "+
				"Please open SORACOM User Console at https://console.soracom.io and check virtual SIM status. "+
				"You have to create arc.json manually", sim.SimId)
		}

		config = &Config{
			PrivateKey:           privateKey,
			PublicKey:            publicKey,
			SimId:                sim.SimId,
			LogLevel:             LogLevelVerbose,
			EnableMetrics:        true,
			Interface:            DefaultInterfaceName(),
			AdditionalAllowedIPs: nil,
			Profile:              b.Profile,
			ArcSession:           &sim.ArcSession,
		}
	} else {
		// or just update arcSession
		arcSession, err := client.CreateArcSession(config.SimId, config.PublicKey.AsWgKey().String())
		if err != nil {
			return nil, err
		}
		config.ArcSession = arcSession
	}

	return config, nil
}
