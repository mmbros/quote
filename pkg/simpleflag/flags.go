package simpleflag

import (
	"fmt"
	"strconv"
	"strings"
)

// String is a flag of string type
type String struct {
	Value  string
	Passed bool
}

// Int is a flag of int type
type Int struct {
	Value  int
	Passed bool
}

// Bool is a flag of bool type
type Bool struct {
	Value  bool
	Passed bool
}

// Strings is a flag of []string type
type Strings []string

func (o *String) String() string {
	return o.Value
}

// Set method of flag.Value interface
func (o *String) Set(value string) error {
	o.Passed = true
	o.Value = value
	return nil
}

func (o *Int) String() string {
	return strconv.Itoa(o.Value)
}

// Set method of flag.Value interface
func (o *Int) Set(value string) (err error) {
	o.Passed = true
	o.Value, err = strconv.Atoi(value)
	return
}

func (o *Bool) String() string {
	return strconv.FormatBool(o.Value)
}

// Set method of flag.Value interface
func (o *Bool) Set(value string) (err error) {
	o.Passed = true
	o.Value, err = strconv.ParseBool(value)
	return
}

// IsBoolFlag method.
// If a Value has an IsBoolFlag() bool method returning true,
// the command-line parser makes -name equivalent to -name=true
// rather than using the next command-line argument.
func (o *Bool) IsBoolFlag() bool { return true }

func (o *Strings) String() string {
	return fmt.Sprintf("[%s]", strings.Join(*o, ", "))
}

// Set method of flag.Value interface
func (o *Strings) Set(value string) error {
	for _, v := range strings.Split(value, ",") {
		*o = append(*o, strings.TrimSpace(v))
	}
	return nil
}
