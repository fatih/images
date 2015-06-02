package command

import (
	"fmt"
	"os"

	"github.com/fatih/flags"
	"github.com/mitchellh/cli"
)

type Delete struct {
	provider string
}

func NewDelete(config *Config) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &Delete{
			provider: config.Provider,
		}, nil
	}
}

func (d *Delete) Help() string {
	if d.provider == "" {
		return `Usage: images delete [options]

  Delete images

Options:

  -provider [name]    Provider to be used to modify images
`
	}

	return Help("delete", d.provider)
}

func (d *Delete) Run(args []string) int {
	if d.provider == "" {
		fmt.Print(d.Help())
		return 1
	}

	if flags.Has("help", args) {
		fmt.Print(d.Help())
		return 1
	}

	p, err := Provider(d.provider, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	deleter, ok := p.(Deleter)
	if !ok {
		err := fmt.Errorf("'%s' doesn't support deleting images", d.provider)
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	if err := deleter.Delete(args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	return 0
}

func (d *Delete) Synopsis() string {
	return "Delete available images"
}
