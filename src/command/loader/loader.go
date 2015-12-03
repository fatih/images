package loader

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
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

// FilterArgs filters the given arguments and returns a filtered argument list.
// It only contains the arguments which are declared in the given configuration
// struct.
func FilterArgs(conf interface{}, args []string) []string {
	configArgs := []string{}

	// need to be declared so we can call it recursively
	var addFields func(fields []*structs.Field)

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

	return configArgs
}

// ExcludeArgs exludes the given arguments declared in the configuration and
// returns the remaining arguments. It's the opposite of FilterArgs
func ExcludeArgs(conf interface{}, args []string) []string {
	// need to be declared so we can call it recursively
	var addFields func(fields []*structs.Field)

	addFields = func(fields []*structs.Field) {
		for _, field := range fields {
			// don't forget nested structs
			if field.Kind() == reflect.Struct {
				addFields(field.Fields())
				continue
			}

			fName := strings.ToLower(strings.Join(camelcase.Split(field.Name()), "-"))
			args = flags.Exclude(fName, args)
		}
	}
	addFields(structs.Fields(conf))

	return args
}

// Load loads the given config to the rules of images CLI
func Load(conf interface{}, args []string) error {
	configArgs := FilterArgs(conf, args)

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

func discoverConfigPath(configName string) (path string, typ string, err error) {
	// Look for a .imagesrc{,.toml,.json} config in current directory first.
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	path, typ, err = discoverConfigPathDir(cwd, configName)
	if err == nil {
		return path, typ, nil
	}
	// Then try the top-level dir of the repository; if we're not in
	// a git repo or there's no git executable, ignore it and
	// return previous error.
	gitTop, gitErr := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if gitErr != nil {
		return "", "", err
	}
	return discoverConfigPathDir(string(bytes.TrimSpace(gitTop)), configName)
}

func discoverConfigPathDir(dir, configName string) (path string, typ string, err error) {
	configPath := filepath.Join(dir, "."+configName)

	if _, err := os.Stat(configPath); err == nil {
		return configPath, "", nil
	}

	// check for .toml
	tomlPath := configPath + ".toml"
	if _, err := os.Stat(tomlPath); err == nil {
		return tomlPath, "toml", nil
	}

	// check for .json
	jsonPath := configPath + ".json"
	if _, err := os.Stat(jsonPath); err == nil {
		return jsonPath, "json", nil
	}

	return "", "", errors.New("couldn't find any .imagesrc file")
}
