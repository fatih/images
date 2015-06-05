package flags

import (
	"flag"
	"reflect"
	"testing"
)

func TestStringList(t *testing.T) {
	f := flag.NewFlagSet("TestTags", flag.PanicOnError)

	var regions []string
	f.Var(StringListVar(&regions), "to", "Regions to be used")
	f.Parse([]string{"-to", "us-east-1,eu-west-2"})

	want := []string{"us-east-1", "eu-west-2"}
	if !reflect.DeepEqual(regions, want) {
		t.Errorf("Regions = %q, want %q", regions, want)
	}
}

func TestIntList(t *testing.T) {
	f := flag.NewFlagSet("TestTags", flag.PanicOnError)

	var ids []int
	f.Var(IntListVar(&ids), "ids", "Ids to be used")
	f.Parse([]string{"-ids", "123,456"})

	want := []int{123, 456}
	if !reflect.DeepEqual(ids, want) {
		t.Errorf("Ids = %q, want %q", ids, want)
	}
}
