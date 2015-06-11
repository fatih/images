package utils

import (
	"io"
	"strings"
	"text/tabwriter"
)

type OutputMode int

const (
	Simplified OutputMode = iota + 1
	JSON
)

func (o OutputMode) String() string {
	switch o {
	case Simplified:
		return "simplified"
	case JSON:
		return "json"
	default:
		return "unknown"
	}

}

var Outputs = map[string]OutputMode{
	"simplified": Simplified,
	"json":       JSON,
}

type OutputValue OutputMode

// NewOutputValue satisfies the flag.Value interface{}. Use it to plug into the
// flag package.
func NewOutputValue(val OutputMode, p *OutputMode) *OutputValue {
	*p = val
	return (*OutputValue)(p)
}

func (o *OutputValue) Set(val string) error {
	// if empty default is used
	if val == "" {
		return nil
	}

	*o = OutputValue(Outputs[strings.ToLower(val)])
	return nil
}

func (o *OutputValue) Get() interface{} { return OutputMode(*o) }

func (o *OutputValue) String() string { return OutputMode(*o).String() }

func NewImagesTabWriter(out io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(out, 10, 8, 1, '\t', 0)
}
