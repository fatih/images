package loader

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/fatih/images/command/flags"
	"github.com/fatih/structs"
	"github.com/koding/multiconfig"
)

var (
	DefaultConfigName = "imagesrc"
)

// Load loads the given config to the rules of images CLI
func Load(conf interface{}, args []string) error {
	// only pass the config's field names arguments
	configArgs := []string{}

	addField := func(field *structs.Field) {
		fieldName := field.Name()
		fName := strings.ToLower(fieldName)

		fmt.Printf("fName = %+v\n", fName)
		val, err := flags.ParseValue(fName, args)
		if err != nil {
			return
		}

		configArgs = append(configArgs, "--"+fName, val)
	}

	for _, field := range structs.Fields(conf) {
		if field.Kind() == reflect.Struct {
			for _, f := range field.Fields() {
				addField(f)
			}
		}

		addField(field)
	}

	fmt.Printf("configArgs = %+v\n", configArgs)

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
		Prefix: "IMAGES",
	}
	f := &multiconfig.FlagLoader{
		Args:      configArgs,
		Flatten:   true,
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
