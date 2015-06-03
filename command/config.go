package command

import (
	"os"

	"github.com/fatih/images/command/loader"
	"github.com/mitchellh/cli"
)

// Config defines the global flag set of images
type Config struct {
	// Provider defines the provider name. It can be a single word or a comma
	// separated list, such as "aws,do". The special "all" name matches all
	// providers.
	Provider string

	// NoColor disables color output
	NoColor bool

	// Force disables asking for user input for certain actions, such as
	// "delete"
	Force bool

	Ui cli.Ui
}

// Help returns the help messages of the respective commands
func (c *Config) Help() map[string]string {
	return map[string]string{
		"provider": "Provider to be used",
		"no-color": "Disables color output",
		"force":    "Disables user prompt",
	}
}

// Load tries to read the global configurations from flag, env or a toml file.
func Load(args []string) (*Config, []string, error) {
	conf := new(Config)
	if err := loader.Load(conf, args); err != nil {
		panic(err)
	}

	remainingArgs := loader.ExcludeArgs(conf, args)

	conf.Ui = &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	return conf, remainingArgs, nil
}
