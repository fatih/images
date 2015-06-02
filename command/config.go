package command

import "github.com/fatih/images/command/loader"

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

// Load tries to read the global configurations from flag, env or a toml file.
func Load(args []string) (*Config, []string, error) {
	conf := new(Config)
	if err := loader.Load(conf, args); err != nil {
		panic(err)
	}

	remainingArgs := loader.ExcludeArgs(conf, args)

	return conf, remainingArgs, nil
}
