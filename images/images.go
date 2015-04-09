package images

import (
	"errors"
	"fmt"
)

type Fetcher interface {
	// Fetch fetches the information from the provider
	Fetch() error

	// Print prints the images to standard output or to something else.
	Print()
}

type Modifier interface {
	Modify(args []string) error
}

type Helper interface {
	Help(command string) string
}

func List(provider string) error {
	p, err := Provider(provider)
	if err != nil {
		return err
	}

	f, ok := p.(Fetcher)
	if !ok {
		return fmt.Errorf("'%s' doesn't support listing images", provider)
	}

	if err := f.Fetch(); err != nil {
		return err
	}

	f.Print()
	return nil
}

func Modify(provider string, args []string) error {
	p, err := Provider(provider)
	if err != nil {
		return err
	}

	m, ok := p.(Modifier)
	if !ok {
		return fmt.Errorf("'%s' doesn't support modifying images", provider)
	}

	return m.Modify(args)
}

func Help(command, provider string) string {
	p, err := Provider(provider)
	if err != nil {
		return "Provider " + provider + " doesn't exists."
	}

	h, ok := p.(Helper)
	if !ok {
		return "No help context available for " + provider
	}

	return h.Help(command)
}

func Provider(provider string) (interface{}, error) {
	switch provider {
	case "aws":
		return NewAwsImages("us-east-1"), nil
	case "digitalocean":
		return nil, errors.New("not supported yet")
	default:
		return nil, errors.New("no such provider available")
	}
}
