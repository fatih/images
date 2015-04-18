package command

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/fatih/images/command/flags"
	"github.com/koding/multiconfig"
)

var (
	envName    = "IMAGES_PROVIDER"
	configName = "imagesrc"
	flagNames  = []string{
		"--provider",
		"--nocolor",
	}
)

type Config struct {
	Provider string
	NoColor  bool
}

// Load tries to read the global configurations from flag, env or a toml file
func Load() (*Config, error) {
	// only pass our flags that are defined in the config, the rest will be
	// handled by the appropriate provider dispatchers
	globalArgs := []string{}
	currentArgs := os.Args[1:]

	for _, fName := range flagNames {
		val, err := flags.ParseValue(fName, currentArgs)
		if err != nil {
			continue
		}

		globalArgs = append(globalArgs, fName, val)
	}

	loader := newImagesLoader(globalArgs)

	conf := new(Config)
	if err := loader.Load(conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func newImagesLoader(args []string) multiconfig.Loader {
	loaders := []multiconfig.Loader{}

	// check for any files
	path, ext, err := discoverConfigPath()
	if err == nil {
		// Choose what while is passed
		switch ext {
		case "json":
			// .imagesrc.json
			loaders = append(loaders, &multiconfig.JSONLoader{Path: path})
		case "toml":
			fallthrough
		default:
			// .imagesrc or .imagesrc.toml
			loaders = append(loaders, &multiconfig.TOMLLoader{Path: path})
		}
	}

	e := &multiconfig.EnvironmentLoader{
		Prefix: "IMAGES",
	}
	f := &multiconfig.FlagLoader{
		Args:      args,
		EnvPrefix: "IMAGES",
	}
	loaders = append(loaders, e, f)

	loader := multiconfig.MultiLoader(loaders...)
	return loader
}

func discoverConfigPath() (string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	configPath := filepath.Join(cwd, "."+configName)

	if _, err := os.Stat(configPath); err == nil {
		return configPath, "", nil
	}

	// check for .toml
	tomlPath := filepath.Join(cwd, "."+configName+".toml")
	if _, err := os.Stat(tomlPath); err == nil {
		return tomlPath, "toml", nil
	}

	// check for .json
	jsonPath := filepath.Join(cwd, "."+configName+".json")
	if _, err := os.Stat(jsonPath); err == nil {
		return jsonPath, "json", nil
	}

	return "", "", errors.New("couldn't find any .imagesrc file")
}
