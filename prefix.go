package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/kingpin"
)

type (
	PrefixValues []PrefixValue

	PrefixValue struct {
		Pattern *regexp.Regexp
		Prefix  string
	}
)

func PrefixValuesFlag(s kingpin.Settings) (target *PrefixValues) {
	target = &PrefixValues{}
	s.SetValue(target)
	return
}

func (v *PrefixValues) Set(value string) error {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("expected pattern:prefix, got '%s'", value)
	}

	pattern, err := regexp.Compile(parts[0])
	if err != nil {
		return fmt.Errorf("expected valid regex pattern, got '%s'", parts[0])
	}

	*v = append(*v, PrefixValue{pattern, parts[1]})
	return nil
}

func (v *PrefixValues) String() string {
	return ""
}
