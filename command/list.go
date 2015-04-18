package command

import (
	"fmt"
	"os"

	"github.com/fatih/images/command/flags"
	"github.com/mitchellh/cli"
)

type List struct {
	provider string
}

func NewList(config *Config) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &List{
			provider: config.Provider,
		}, nil
	}
}

func (l *List) Help() string {
	if l.provider == "" {
		return `Usage: images list [options]

  Lists available images for the given provider.

Options:

  -provider [name]    Provider to be used to modify images
`
	}

	return Help("list", l.provider)
}

func (l *List) Run(args []string) int {
	if l.provider == "" {
		fmt.Print(l.Help())
		return 1
	}

	remainingArgs := flags.FilterFlag("provider", args)

	p, err := Provider(l.provider, remainingArgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	f, ok := p.(Fetcher)
	if !ok {
		err := fmt.Errorf("'%s' doesn't support listing images", l.provider)
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	if err := f.Fetch(args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	f.Print()
	return 0
}

func (l *List) Synopsis() string {
	return "List available images"
}
