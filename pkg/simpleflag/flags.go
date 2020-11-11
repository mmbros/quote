package simpleflag

import (
	"fmt"
	"strconv"
	"strings"
)

// String is a flag of string type
type String struct {
	value  string
	passed bool
}

// Int is a flag of int type
type Int struct {
	value  int
	passed bool
}

// Bool is a flag of bool type
type Bool struct {
	value  bool
	passed bool
}

// Strings is a flag of []string type
type Strings []string

func (o *String) String() string {
	return o.value
}

// Set method of flag.Value interface
func (o *String) Set(value string) error {
	o.passed = true
	o.value = value
	return nil
}

func (o *Int) String() string {
	return strconv.Itoa(o.value)
}

// Set method of flag.Value interface
func (o *Int) Set(value string) (err error) {
	o.passed = true
	o.value, err = strconv.Atoi(value)
	return
}

func (o *Bool) String() string {
	return strconv.FormatBool(o.value)
}

// Set method of flag.Value interface
func (o *Bool) Set(value string) (err error) {
	o.passed = true
	o.value, err = strconv.ParseBool(value)
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
