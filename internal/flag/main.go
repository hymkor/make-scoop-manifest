package flag

// On the standard version of "flag", parsing stops just before the first
// non-flag argument.
//
// On this version, it continues for compatiblity for v0.9.0

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var Debug io.Writer = io.Discard

//var Debug io.Writer = os.Stderr

type _StringFlag struct {
	_name    string
	_default string
	_usage   string
	_value   string
}

func (s *_StringFlag) parse(args []string, log io.Writer) ([]string, error) {
	if len(args) <= 0 {
		return nil, errors.New("too few arguments")
	}
	s._value = args[0]
	fmt.Fprintf(Debug, "%s: set %#v\n", s._name, s._value)
	return args[1:], nil
}

func (s *_StringFlag) usage() string {
	var u strings.Builder
	fmt.Fprintf(&u, "  %s string", s._name)
	if u.Len() <= 6 {
		u.WriteByte('\t')
	} else {
		u.WriteString("\n    \t")
	}
	u.WriteString(s._usage)
	return u.String()
}

type _BoolFlag struct {
	_name    string
	_default bool
	_usage   string
	_value   bool
}

func (b *_BoolFlag) parse(args []string, log io.Writer) ([]string, error) {
	b._value = true
	fmt.Fprintf(Debug, "%s: set %#v\n", b._name, b._value)
	return args, nil
}

func (b *_BoolFlag) usage() string {
	var u strings.Builder
	fmt.Fprintf(&u, "  %s", b._name)
	if u.Len() <= 6 {
		u.WriteByte('\t')
	} else {
		u.WriteString("\n    \t")
	}
	u.WriteString(b._usage)
	return u.String()
}

type _Flag interface {
	parse([]string, io.Writer) ([]string, error)
	usage() string
}

var flags = map[string]_Flag{}

func String(name, defaults, usage string) *string {
	o := &_StringFlag{
		_name:    "-" + name,
		_default: defaults,
		_usage:   usage,
		_value:   defaults,
	}
	flags[name] = o
	return &o._value
}

func Bool(name string, defaults bool, usage string) *bool {
	b := &_BoolFlag{
		_name:    "-" + name,
		_default: defaults,
		_usage:   usage,
		_value:   defaults,
	}
	flags[name] = b
	return &b._value
}

var nonOptions = []string{}

func Args() []string {
	return nonOptions
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	for _, value := range flags {
		fmt.Fprintln(os.Stderr, value.usage())
	}
}

func parse() error {
	args := os.Args[1:]
	for len(args) > 0 {
		name := args[0]
		args = args[1:]
		if len(name) <= 0 || name[0] != '-' {
			nonOptions = append(nonOptions, name)
			continue
		}
		o, ok := flags[name[1:]]
		if !ok {
			if name == "-h" {
				usage()
				os.Exit(0)
			}
			return fmt.Errorf("flag provided but not defined: %s", name)
		}
		var err error
		args, err = o.parse(args, os.Stderr)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	fmt.Fprintf(Debug, "Non Option args: %#v\n", nonOptions)
	return nil
}

func Parse() {
	err := parse()
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err.Error())
	usage()
	os.Exit(1)
}
