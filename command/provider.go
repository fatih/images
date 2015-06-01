package command

import (
	"errors"
	"fmt"

	"github.com/fatih/images/provider/awsimages"
	"github.com/fatih/images/provider/doimages"
	"github.com/fatih/images/provider/gceimages"
)

var errNoProvider = errors.New("no such provider available")

// Provider returns the provider with the given name
func Provider(name string, args []string) (interface{}, error) {
	switch name {
	case "aws":
		return awsimages.New(args)
	case "do":
		return doimages.New(args)
	case "gce":
		return gceimages.New(args)
	default:
		return nil, errNoProvider
	}
}

// Fecher fetches and prints the images
type Fetcher interface {
	// Fetch fetches the information from the provider
	Fetch(args []string) error

	// Print prints the images to standard output or to something else.
	Print()
}

// Copyier copyies the image.
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
		if err == errNoProvider {
			return "Provider '" + name + "' doesn't exists."
		}

		// provider exists but failed, print the help for it
		fmt.Println(err, "\n")
	}

	h, ok := p.(Helper)
	if !ok {
		return "No help context available for " + name
	}

	return h.Help(command)
}
