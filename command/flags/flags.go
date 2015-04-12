// Package flags is low level package for parsing single flag arguments and
// their associated values. It's useful for CLI applications or applications
// that parses os.Args manually.
package flags

import (
	"errors"
	"fmt"
)

// HasFlag checks whether the given flag is available or not in the argument
// list
func HasFlag(args []string) bool {
	for _, arg := range args {
		if _, err := ParseFlag(arg); err == nil {
			return true
		}
	}

	return false
}

// IsFlag checks whether the given argument is a valid flag or not
func IsFlag(arg string) bool {
	if _, err := ParseFlag(arg); err != nil {
		return false
	}

	return true
}

// ParseFlag parses a flags name. A flag can be in form of --name=value,
// -name=value, -n=value, or --name, -name=, etc...  If it's a correct flag,
// the name is returned. If not an empty string and an error message is
// returned
func ParseFlag(arg string) (string, error) {
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

// ParseValue parses the value from the given flag. A flag name can be in
// form of name=value, n=value, n=, n.
func ParseValue(flag string) (name, value string) {
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

// ParseName parses the given flagName from the args slice and returns the
// value passed to the flag. An example: args: ["--provider", "aws"] will
// return "aws".
func ParseFlagValue(flagName string, args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("argument slice is empty")
	}

	for i, arg := range args {
		flag, err := ParseFlag(arg)
		if err != nil {
			continue
		}

		name, value := ParseValue(flag)
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
			if !IsFlag(args[i+1]) {
				return args[i+1], nil
			}
		}
	}

	return "", fmt.Errorf("argument is not passed to flag: %s", flagName)
}

// FilterFlag filters the given valid flagName  with it's associated value (or
// none) from the args. It returns the remaining arguments. If no flagName is
// passed or if the flagName is invalid, remaining arguments are returned
// without any change.
func FilterFlag(flagName string, args []string) []string {
	if len(args) == 0 {
		return args
	}

	for i, arg := range args {
		flag, err := ParseFlag(arg)
		if err != nil {
			continue
		}

		name, value := ParseValue(flag)
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

			// flag is between the first and the last, delete it and return the
			// remaining arguments
			return append(args[:i], args[i+1:]...)
		}

		// no value found yet, check out the next argument. at least two args
		// must be present
		if len(args) < i+1 {
			continue
		}

		// only one flag is passed and it's ours in the form of ["--flagName"]
		if len(args) == 1 {
			return args[1:]
		}

		// flag is the latest item and has no value, return til the flagName,
		// ["--foo", "bar", "--flagName"]
		if len(args) == i+1 {
			return args[:i]
		}

		// next argument is a flag i.e: "--flagName --otherFlag", remove our
		// flag and return the remainings
		if IsFlag(args[i+1]) {
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
