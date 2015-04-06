package command

import "testing"

func TestParseFlag(t *testing.T) {
	var flags = []struct {
		flag string
		name string
	}{
		{name: "foo", flag: "--foo"},
		{name: "foo", flag: "-foo"},
		{name: "foo=bar", flag: "-foo=bar"},
		{name: "foo=", flag: "-foo="},
		{name: "foo=b", flag: "-foo=b"},
		{name: "f", flag: "-f"},
		{name: "f", flag: "--f"},
		{name: "", flag: "---f"},
		{name: "", flag: "f"},
		{name: "", flag: "--"},
		{name: "", flag: "-"},
	}

	for _, f := range flags {
		name, _ := parseFlag(f.flag)
		if name != f.name {
			t.Errorf("flag: %s\n\twant: %s\n\tgot : %s\n", f.flag, f.name, name)
		}
	}

}

func TestParseValue(t *testing.T) {
	var flags = []struct {
		flag  string
		value string
	}{
		{flag: "foo=bar", name: "foo", value: "bar"},
		{flag: "foo=b", name: "foo", value: "b"},
		{flag: "f=", name: "f", value: ""},
		{flag: "f", name: "f", value: ""},
		{flag: "", name: "", value: ""},
	}

	for _, f := range flags {
		name, value := parseValue(f.flag)
		if value != f.value {
			t.Errorf("flag: %s\n\twant: %s\n\tgot : %s\n", f.flag, f.value, value)
		}

		if name != f.name {
			t.Errorf("flag: %s\n\twant: %s\n\tgot : %s\n", f.flag, f.name, name)
		}
	}

}
