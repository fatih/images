package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fatih/images/images"
	"github.com/mitchellh/cli"
)

type Modify struct{}

func NewModify() (cli.Command, error) {
	return &Modify{}, nil
}

func (m *Modify) Help() string {
	return `Usage: images modify [options] 

  Modifies images properties. Each providers sub options are different.

Options:

  -provider                  Provider to be used to modify images
`
}

func (m *Modify) Run(args []string) int {
	var (
		provider string
	)

	flagSet := flag.NewFlagSet("modify", flag.ContinueOnError)
	flagSet.StringVar(&provider, "provider", "", "cloud provider to modify images")
	flagSet.SetOutput(ioutil.Discard)

	if err := flagSet.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	fmt.Printf("provider = %+v\n", provider)

	if flagSet.NFlag() == 0 {
		fmt.Print(m.Help())
		return 1
	}

	if err := images.Modify(provider, args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	return 0
}

func (m *Modify) Synopsis() string {
	return "Modify image properties"
}
