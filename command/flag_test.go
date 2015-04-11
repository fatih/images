package command

import (
	"reflect"
	"testing"
)

func TestIsFlag(t *testing.T) {
	var flags = []struct {
		flag   string
		isFlag bool
	}{
		{flag: "--foo", isFlag: true},
		{flag: "--foo=bar", isFlag: true},
		{flag: "-foo", isFlag: true},
		{flag: "-foo=bar", isFlag: true},
		{flag: "-f", isFlag: true},
		{flag: "-f=bar", isFlag: true},
		{flag: "f=bar", isFlag: false},
		{flag: "f=", isFlag: false},
		{flag: "f", isFlag: false},
		{flag: "", isFlag: false},
	}

	for _, f := range flags {
		is := isFlag(f.flag)
		if is != f.isFlag {
			t.Errorf("flag: %s\n\twant: %s\n\tgot : %s\n", f.flag, f.isFlag, is)
		}
	}
}

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
		name  string
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
			t.Errorf("parsing value from flag: %s\n\twant: %s\n\tgot : %s\n",
				f.flag, f.value, value)
		}

		if name != f.name {
			t.Errorf("parsing name from flag: %s\n\twant: %s\n\tgot : %s\n",
				f.flag, f.name, name)
		}
	}

}

func TestParseFlagValue(t *testing.T) {
	var arguments = []struct {
		args  []string
		value string
	}{
		{args: []string{"--provider=aws", "foo"}, value: "aws"},
		{args: []string{"-provider=aws", "foo", "bar"}, value: "aws"},
		{args: []string{"-provider=aws,do"}, value: "aws,do"},
		{args: []string{"-p=aws", "aws"}, value: "aws"},
		{args: []string{"--provider", "aws"}, value: "aws"},
		{args: []string{"-provider", "aws"}, value: "aws"},
		{args: []string{"-p", "aws"}, value: "aws"},
		{args: []string{"-p"}, value: ""},
		{args: []string{"-p="}, value: ""},
		{args: []string{"-p=", "--foo"}, value: ""},
		{args: []string{"-p", "--foo"}, value: ""},
		{args: []string{"-provider", "--foo"}, value: ""},
		{args: []string{"--provider", "--foo"}, value: ""},
	}

	for _, args := range arguments {
		value, _ := parseFlagValue("provider", args.args)

		if value != args.value {
			t.Errorf("parsing args value: %v\n\twant: %s\n\tgot : %s\n",
				args.args, args.value, value)
		}
	}
}

func TestFilterFlag(t *testing.T) {
	var arguments = []struct {
		args    []string
		remArgs []string
	}{
		{args: []string{}, remArgs: []string{}},
		{args: []string{"-provider"}, remArgs: []string{}},
		{args: []string{"--provider"}, remArgs: []string{}},
		{args: []string{"--provider=aws", "foo"}, remArgs: []string{"foo"}},
		{args: []string{"-provider=aws", "foo", "bar"}, remArgs: []string{"foo", "bar"}},
		{args: []string{"-provider=aws,do"}, remArgs: []string{}},
		{args: []string{"-p=aws", "aws"}, remArgs: []string{"aws"}},
		{args: []string{"--test", "foo", "-p=aws", "aws"}, remArgs: []string{"--test", "foo", "aws"}},
		{args: []string{"--test", "foo", "--provider=aws", "foo"}, remArgs: []string{"--test", "foo", "foo"}},
		{args: []string{"--example", "foo"}, remArgs: []string{"--example", "foo"}},
		{args: []string{"--test", "--provider", "aws"}, remArgs: []string{"--test"}},
		{args: []string{"--test", "--provider", "aws", "--test2"}, remArgs: []string{"--test", "--test2"}},
		{args: []string{"--test", "bar", "--provider", "aws"}, remArgs: []string{"--test", "bar"}},
		{args: []string{"--provider", "aws"}, remArgs: []string{}},
		{args: []string{"--provider", "aws", "--test"}, remArgs: []string{"--test"}},
		{args: []string{"--provider", "--test"}, remArgs: []string{"--test"}},
		{args: []string{"--test", "--provider"}, remArgs: []string{"--test"}},
		{args: []string{"--test", "bar", "--foo", "--provider"}, remArgs: []string{"--test", "bar", "--foo"}},
		{args: []string{"--test", "--provider", "--test2", "aws"}, remArgs: []string{"--test", "--test2", "aws"}},
	}

	for _, args := range arguments {
		remainingArgs := filterFlag("provider", args.args)

		if !reflect.DeepEqual(remainingArgs, args.remArgs) {
			t.Errorf("parsing and returning rem args: %v\n\twant: %s\n\tgot : %s\n",
				args.args, args.remArgs, remainingArgs)
		}
	}
}
