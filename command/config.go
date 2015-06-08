package command

import (
	"os"

	"github.com/fatih/images/command/loader"
	"github.com/mitchellh/cli"
)

// Config defines the global flag set of images
type Config struct {
	// Providers define the providers to be used with images. Example: ["aws",
	// "do"]. The special ["all"] name matches all providers.
	Providers []string `toml:"providers" json:"providers"`

	// NoColor disables color output
	NoColor bool `toml:"no_color" json:"no_color"`

	// Force disables asking for user input for certain actions, such as
	// "delete"
	Force bool `toml:"force" json:"force"`

	Ui cli.Ui `toml:"-" json:"-"`
}

// Help returns the help messages of the respective commands
func (c *Config) Help() map[string]string {
	return map[string]string{
		"providers": "Providers to be used",
		"no-color":  "Disables color output",
		"force":     "Disables user prompt",
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
