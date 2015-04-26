package loader

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/fatih/flags"
	"github.com/fatih/structs"
	"github.com/koding/multiconfig"
)

var (
	DefaultConfigName = "imagesrc"
)

// Load loads the given config to the rules of images CLI
func Load(conf interface{}, args []string) error {
	configArgs := []string{}

	// need to be declared so we can call it recursively :)
	var addFields func(fields []*structs.Field)

	// only pass the config's field names arguments. This means we create an
	// argument list that contains flag names and their values which are only
	// declared in the passed "conf" struct. Any other argument is discarded
	addFields = func(fields []*structs.Field) {
		for _, field := range fields {
			// don't forget nested structs
			if field.Kind() == reflect.Struct {
				addFields(field.Fields())
				continue
			}

			fName := strings.ToLower(strings.Join(camelcase.Split(field.Name()), "-"))
			val, err := flags.Value(fName, args)
			if err != nil {
				continue
			}

			configArgs = append(configArgs, "--"+fName, val)
		}
	}
	addFields(structs.Fields(conf))

	loaders := []multiconfig.Loader{}

	// check for any files
	path, ext, err := discoverConfigPath(DefaultConfigName)
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
		Prefix:    "IMAGES",
		CamelCase: true,
	}
	f := &multiconfig.FlagLoader{
		Args:      configArgs,
		Flatten:   true,
		CamelCase: true,
		EnvPrefix: "IMAGES",
	}
	loaders = append(loaders, e, f)

	l := multiconfig.MultiLoader(loaders...)
	return l.Load(conf)
}

func discoverConfigPath(configName string) (string, string, error) {
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
