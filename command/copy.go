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
	if len(c.Providers) != 1 {
		return `Usage: images copy [options]

  Copy images to regions

Options:

  -providers "name"    Provider to be used to copy images
`
	}

	return Help("copy", c.Providers[0])
}

func (c *Copy) Run(args []string) int {
	if len(c.Providers) != 1 {
		fmt.Print(c.Help())
		return 1
	}

	if flags.Has("help", args) {
		fmt.Print(c.Help())
		return 1
	}

	provider := c.Providers[0]
	if provider == "all" {
		fmt.Fprintln(os.Stderr, "Copy doesn't support multiple providers")
		return 1
	}

	p, remArgs, err := Provider(provider, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	copier, ok := p.(Copier)
	if !ok {
		err := fmt.Errorf("'%s' doesn't support copying images", provider)
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	if err := copier.Copy(remArgs); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	return 0
}

func (c *Copy) Synopsis() string {
	return "Copy/transfer images"
}
