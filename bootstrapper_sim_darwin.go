package soratun

import (
	"errors"
)

// SimBootstrapper defines bootstrap method with SORACOM Krypton SIM authentication. Needs krypton-cli installed.
type SimBootstrapper struct {
	KryptonCliPath string
	Arguments      []string
}

// Execute calls SORACOM Krypton CLI to create a new virtual subscriber which is associated with current physical SIM.
func (b *SimBootstrapper) Execute(config *Config) (*Config, error) {
	return nil, errors.New("bootstrap with SIM authentication is not supported on this platform")
}
