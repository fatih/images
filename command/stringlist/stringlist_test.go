package stringlist

import (
	"flag"
	"reflect"
	"testing"
)

func TestTags(t *testing.T) {
	f := flag.NewFlagSet("TestTags", flag.PanicOnError)

	var regions []string
	f.Var(New([]string{}, &regions), "to", "Regions to be used")
	f.Parse([]string{"-to", "us-east-1,eu-west-2"})

	want := []string{"us-east-1", "eu-west-2"}
	if !reflect.DeepEqual(regions, want) {
		t.Errorf("Regions = %q, want %q", regions, want)
	}
}

func TestTagsDefault(t *testing.T) {
	f := flag.NewFlagSet("TestTags", flag.PanicOnError)

	var regions []string
	var defaults = []string{"us-east-1"}
	f.Var(New(defaults, &regions), "to", "Regions to be used")
	f.Parse([]string{"-to", ""})

	if !reflect.DeepEqual(regions, defaults) {
		t.Errorf("Regions = %q, want %q", regions, defaults)
	}
}
