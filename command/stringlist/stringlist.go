// Package stringlist implements a flag that accepts a comma separated string
// which converts to a slice containing the elements.
package stringlist

import "strings"

// StringListValue is an implementation of flag.Value interface that parses
// accepts a comma separated list of elements and populates a []string slice.
//
// Example:
//
//  var regions []string
//	flag.Var(stringutils.New([]string{"default"}, &regions), "to", "Regions to be used")
type StringListValue []string

// New returns a new *StringListValue. Use it flag.Var() or flagset.Var()
func New(val []string, p *[]string) *StringListValue {
	*p = val
	return (*StringListValue)(p)
}

func (s *StringListValue) Set(val string) error {
	// if empty default is used
	if val == "" {
		return nil
	}

	*s = StringListValue(strings.Split(val, ","))
	return nil
}

func (s *StringListValue) Get() interface{} { return []string(*s) }

func (s *StringListValue) String() string { return strings.Join(*s, ",") }
