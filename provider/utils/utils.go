package utils

import "strings"

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
