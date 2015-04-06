package command

import (
	"errors"
	"fmt"
)

// parseFlag parses a flags name. A flag can be in form of --name, -name or -n.
// If it's a correct flag, the name is returned. If not an empty string and an
// error message is returned
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

	return
}

func parseProvider(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("argument is empty")
	}

	for _, arg := range args {
		name, err := parseFlag(arg)
		if err != nil {
			fmt.Println("err")
			continue
		}

		val := parseValue(name)

		fmt.Printf("val = %+v\n", val)

		fmt.Printf("name = %+v\n", name)
	}

	return "", nil
}
