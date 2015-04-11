package command

import (
	"errors"
	"fmt"
)

// isFlag checks whether the given argument is a valid flag or not
func isFlag(arg string) bool {
	if _, err := parseFlag(arg); err != nil {
		return false
	}

	return true
}

// parseFlag parses a flags name. A flag can be in form of --name=value,
// -name=value, -n=value, or --name, -name=, etc...  If it's a correct flag,
// the name is returned. If not an empty string and an error message is
// returned
func parseFlag(arg string) (string, error) {
	if arg == "" {
		return "", errors.New("argument is empty")
	}

	if len(arg) == 1 {
		return "", errors.New("argument is too short")
	}

	if arg[0] != '-' {
		return "", errors.New("argument doesn't start with dash")
	}

	numMinuses := 1

	if arg[1] == '-' {
		numMinuses++
		if len(arg) == 2 {
			return "", errors.New("argument is too short")
		}
	}

	name := arg[numMinuses:]
	if len(name) == 0 || name[0] == '-' || name[0] == '=' {
		return "", fmt.Errorf("bad flag syntax: %s", arg)
	}

	return name, nil
}

// parseValue parses the value from the given flag. A flag name can be in
// form of name=value, n=value, n=, n.
func parseValue(flag string) (name, value string) {
	for i, r := range flag {
		if r == '=' {
			value = flag[i+1:]
			name = flag[0:i]
		}
	}

	// special case of "n"
	if name == "" {
		name = flag
	}

	return
}

// parseName parses the given flagName from the args slice and returns the
// value passed to the flag. An example: args: ["--provider", "aws"], flagName:
// "provider" will return "aws" and ["--foo"].
func parseFlagValue(flagName string, args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("argument slice is empty")
	}

	for i, arg := range args {
		flag, err := parseFlag(arg)
		if err != nil {
			continue
		}

		name, value := parseValue(flag)
		if name != flagName && name[0] != flagName[0] {
			continue
		}

		if value != "" {
			return value, nil // found
		}

		// no value found, check out the next argument. at least two args must
		// be present
		if len(args) > i+1 {
			// value must be next argument
			if !isFlag(args[i+1]) {
				return args[i+1], nil
			}
		}
	}

	return "", fmt.Errorf("argument is not passed to flag: %s", flagName)
}

// filterFlag filters the given valid flagName  with it's associated value (or
// none) from the args. It returns the remaining arguments. If no flagName is
// passed or if the flagName is invalid, remaining arguments are returned
// without any change.
func filterFlag(flagName string, args []string) []string {
	if len(args) == 0 {
		return args
	}

	for i, arg := range args {
		flag, err := parseFlag(arg)
		if err != nil {
			continue
		}

		name, value := parseValue(flag)
		if name != flagName && name[0] != flagName[0] {
			continue
		}

		// flag is in the form of "--flagName=value"
		if value != "" {
			// our flag is the first item in the argument list, so just return
			// the remainings
			if i <= 1 {
				return args[i+1:]
			}

			// flag is between the first and the last, delete and return the
			// remaining arguments
			return append(args[:i], args[i+1:]...)
		}

		// no value found, check out the next argument. at least two args must
		// be present
		if len(args) < i+1 {
			continue
		}

		// next argument is a flag i.e: "--flagName --otherFlag", remove our
		// flag and return the remainings
		if isFlag(args[i+1]) {
			// flag is between the first and the last, delete and return the
			// remaining arguments
			return append(args[:i], args[i+1:]...)
		}

		// next flag is a value, +2 because the flag is in the form of
		// "--flagName value".  This means we need to remove two items from the
		// slice
		return append(args[:i], args[i+2:]...)
	}

	return args // nothing found
}
