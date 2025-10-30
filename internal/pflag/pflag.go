// Package pflag provides command line flag parsing
package pflag

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// StringSlice is a custom flag type for comma-separated strings
type StringSlice []string

func (s *StringSlice) String() string {
	return strings.Join(*s, ",")
}

func (s *StringSlice) Set(value string) error {
	if value == "" {
		*s = []string{}
		return nil
	}
	*s = strings.Split(value, ",")
	return nil
}