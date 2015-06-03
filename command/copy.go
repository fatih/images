package command

import (
	"fmt"
	"os"

	"github.com/fatih/flags"
	"github.com/mitchellh/cli"
)

type Copy struct {
	*Config
}

func NewCopy(config *Config) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &Copy{
			Config: config,
		}, nil
	}
}

func (c *Copy) Help() string {
	if c.Provider == "" {
		return `Usage: images copy [options]

  Copy images to regions

Options:

  -provider [name]    Provider to be used to modify images
`
	}

	return Help("copy", c.Provider)
}

func (c *Copy) Run(args []string) int {
	if c.Provider == "" {
		fmt.Print(c.Help())
		return 1
	}

	if flags.Has("help", args) {
		fmt.Print(c.Help())
		return 1
	}

	p, err := Provider(c.Provider, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	copyier, ok := p.(Copyier)
	if !ok {
		err := fmt.Errorf("'%s' doesn't support copying images", c.Provider)
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
	return "Copy/transfer images"
}
