package command

import (
	"os"

	"github.com/fatih/images/command/loader"
)

// Config defines the global flag set of images
type Config struct {
	Provider string
	NoColor  bool
}

// Help returns the help messages of the respective commands
func (c *Config) Help() map[string]string {
	return map[string]string{
		"provider": "Provider to be used",
		"no-color": "No color disables color output",
	}
}

// Load tries to read the global configurations from flag, env or a toml file
func Load() (*Config, error) {
	// only pass our flags that are defined in the config, the rest will be
	// handled by the appropriate provider dispatchers
	currentArgs := os.Args[1:]

	conf := new(Config)
	if err := loader.Load(conf, currentArgs); err != nil {
		panic(err)
	}

	return conf, nil
}
