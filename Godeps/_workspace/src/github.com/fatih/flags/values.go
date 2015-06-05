package flags

import (
	"strconv"
	"strings"
)

// StringList is an implementation of flag.Value interface that accepts a comma
// separated list of elements and populates a []string slice.
//
// Example:
//
//  var regions []string
//  flag.Var(flags.StringListVar(&regions), "to", "Regions to be used")
type StringList []string

// New returns a new *StringList.
func StringListVar(p *[]string) *StringList {
	return (*StringList)(p)
}

func (s *StringList) Set(val string) error {
	// if empty default is used
	if val == "" {
		return nil
	}

	*s = StringList(strings.Split(val, ","))
	return nil
}

func (s *StringList) Get() interface{} { return []string(*s) }

func (s *StringList) String() string { return strings.Join(*s, ",") }

// IntList is an implementation of flag.Value interface that accepts a comma
// separated list of integers and populates a []int slice.
//
// Example:
//
//  var ids []int
//  flag.Var(flags.IntListVar(&ids), "server", "IDs to be used")
type IntList []int

// New returns a new *IntList.
func IntListVar(p *[]int) *IntList {
	return (*IntList)(p)
}

func (i *IntList) Set(val string) error {
	// if empty default is used
	if val == "" {
		return nil
	}

	var list []int
	for _, in := range strings.Split(val, ",") {
		i, err := strconv.Atoi(in)
		if err != nil {
			return err
		}

		list = append(list, i)
	}

	*i = IntList(list)
	return nil
}

func (i *IntList) Get() interface{} { return []int(*i) }

func (i *IntList) String() string {
	var list []string
	for _, in := range *i {
		list = append(list, strconv.Itoa(in))
	}
	return strings.Join(list, ",")
}
