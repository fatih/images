package command

import "os"

var (
	EnvName = "IMAGES_PROVIDER"
)

func providerFromEnvOrFlag(args []string) (string, error) {
	// first read from env
	p := os.Getenv(EnvName)
	if p != "" {
		return p, nil
	}

	// second from flag
	p, err := parseFlagValue("provider", args)
	if err != nil {
		return "", err
	}

	return p, nil
}
