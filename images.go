package images

import "errors"

type Config struct {
	Provider string
}

type Image struct {
	ID string
}

type ImageProvider interface {
	Fetch() error
	Print()
}

func Run(conf *Config) error {
	i, err := Provider(conf.Provider)
	if err != nil {
		return err
	}

	if err := i.Fetch(); err != nil {
		return err
	}

	i.Print()

	return nil
}

func Provider(provider string) (ImageProvider, error) {
	switch provider {
	case "aws":
		return NewAwsImages("us-east-1"), nil
	case "digitalocean":
		return nil, errors.New("not supported yet")
	default:
		return nil, errors.New("no such provider available")
	}
}
