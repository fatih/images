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
	if len(m.Providers) != 1 {
		defaultHelp := `Usage: images modify [options]

  Modifies images properties. Each providers sub options are different.

Options:

  -providers                  Provider to be used to modify images
`
		return defaultHelp
	}

	return Help("modify", m.Providers[0])
}

func (m *Modify) Run(args []string) int {
	if len(m.Providers) != 1 {
		fmt.Print(m.Help())
		return 1
	}

	if flags.Has("help", args) {
		fmt.Print(m.Help())
		return 1
	}

	provider := m.Providers[0]
	if provider == "all" {
		fmt.Fprintln(os.Stderr, "Modify doesn't support multiple providers")
		return 1
	}

	p, err := Provider(provider, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	mr, ok := p.(Modifier)
	if !ok {
		err := fmt.Errorf("'%s' doesn't support listing images", provider)
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
