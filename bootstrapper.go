package soratun

// Bootstrapper defines how to bootstrap virtual SIM with SORACOM.
type Bootstrapper interface {
	Execute(config *Config) (*Config, error)
}
