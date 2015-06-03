package command

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fatih/flags"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/cli"
)

type List struct {
	*Config
}

func NewList(config *Config) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &List{
			Config: config,
		}, nil
	}
}

func (l *List) Help() string {
	if l.Provider == "" {
		return `Usage: images list [options]

  Lists available images for the given provider.

Options:

  -provider "name,..."    Provider to be used to modify images
`
	}

	return Help("list", l.Provider)
}

func (l *List) Run(args []string) int {
	if l.Provider == "" {
		fmt.Println(l.Help())
		return 1
	}

	if flags.Has("help", args) {
		fmt.Println(l.Help())
		return 1
	}

	providers := strings.Split(l.Provider, ",")
	if l.Provider == "all" {
		providers = providerList
	}

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex // protects multiErrors
		multiErrors error
	)

	printProvider := func(provider string) error {
		p, err := Provider(provider, args)
		if err != nil {
			if err == errNoProvider {
				return errors.New("Provider '" + provider + "' doesn't exists.")
			}
		}

		f, ok := p.(Fetcher)
		if !ok {
			return fmt.Errorf("Provider '%s' doesn't support listing images", l.Provider)
		}

		if err := f.Fetch(args); err != nil {
			// we don't return here, because Print might display at least
			// successfull results.
			fmt.Fprintln(os.Stderr, err.Error())
		}

		f.Print()
		fmt.Println("")
		return nil
	}

	for _, provider := range providers {
		wg.Add(1)
		go func(provider string) {
			err := printProvider(provider)
			if err != nil {
				mu.Lock()
				multiErrors = multierror.Append(multiErrors, err)
				mu.Unlock()
			}
			wg.Done()
		}(provider)
	}

	wg.Wait()

	if multiErrors != nil {
		fmt.Fprintln(os.Stderr, multiErrors.Error())
		return 1
	}

	return 0
}

func (l *List) Synopsis() string {
	return "List available images"
}
