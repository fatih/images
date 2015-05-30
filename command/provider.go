package command

import (
	"errors"

	"github.com/fatih/images/provider/awsimages"
	"github.com/fatih/images/provider/doimages"
)

// Provider returns the provider with the given name
func Provider(name string, args []string) (interface{}, error) {
	switch name {
	case "aws":
		return awsimages.New(args), nil
	case "do":
		return doimages.New(args), nil
	default:
		return nil, errors.New("no such provider available")
	}
}

// Fecher fetches and prints the
type Fetcher interface {
	// Fetch fetches the information from the provider
	Fetch(args []string) error

	// Print prints the images to standard output or to something else.
	Print()
}

type Copyier interface {
	Copy(args []string) error
}

// Deleter delets images
type Deleter interface {
	Delete(args []string) error
}

// Modifier modifies image attributes
type Modifier interface {
	Modify(args []string) error
}

// Helper returns the help message
type Helper interface {
	Help(command string) string
}

func Help(command, name string) string {
	p, err := Provider(name, nil)
	if err != nil {
		return "Provider " + name + " doesn't exists."
	}

	h, ok := p.(Helper)
	if !ok {
		return "No help context available for " + name
	}

	return h.Help(command)
}
