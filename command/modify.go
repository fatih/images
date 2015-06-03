package command

import (
	"fmt"
	"os"

	"github.com/fatih/flags"
	"github.com/mitchellh/cli"
)

type Modify struct {
	*Config
}

func NewModify(config *Config) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &Modify{
			Config: config,
		}, nil
	}
}

func (m *Modify) Help() string {
	if m.Provider == "" {
		defaultHelp := `Usage: images modify [options]

  Modifies images properties. Each providers sub options are different.

Options:

  -provider                  Provider to be used to modify images
`
		return defaultHelp
	}

	return Help("modify", m.Provider)
}

func (m *Modify) Run(args []string) int {
	if m.Provider == "" {
		fmt.Print(m.Help())
		return 1
	}

	if flags.Has("help", args) {
		fmt.Print(m.Help())
		return 1
	}

	p, err := Provider(m.Provider, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	mr, ok := p.(Modifier)
	if !ok {
		err := fmt.Errorf("'%s' doesn't support listing images", m.Provider)
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	if err := mr.Modify(args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	return 0
}

func (m *Modify) Synopsis() string {
	return "Modify image properties"
}
