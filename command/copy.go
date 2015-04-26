package command

import (
	"fmt"
	"os"

	"github.com/fatih/flags"
	"github.com/mitchellh/cli"
)

type Copy struct {
	provider string
}

func NewCopy(config *Config) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &Copy{
			provider: config.Provider,
		}, nil
	}
}

func (c *Copy) Help() string {
	if c.provider == "" {
		return `Usage: images copy [options]

  Copy images to different regions

Options:

  -provider [name]    Provider to be used to modify images
`
	}

	return Help("copy", c.provider)
}

func (c *Copy) Run(args []string) int {
	if c.provider == "" {
		fmt.Print(c.Help())
		return 1
	}

	if flags.Has("help", args) {
		fmt.Print(c.Help())
		return 1
	}

	remainingArgs := flags.Exclude("provider", args)

	p, err := Provider(c.provider, remainingArgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	copyier, ok := p.(Copyier)
	if !ok {
		err := fmt.Errorf("'%s' doesn't support copying images", c.provider)
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	if err := copyier.Copy(args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	return 0
}

func (c *Copy) Synopsis() string {
	return "Copy images to different regions"
}
