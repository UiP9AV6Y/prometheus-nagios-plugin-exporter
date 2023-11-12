package config

import (
	"bytes"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

const (
	trueString  = BoolString("true")
	falseString = BoolString("false")
)

// Argument defines the condition and representation of a commandline
// argument to a module command
type Argument struct {
	Condition BoolString `yaml:"set_if,omitempty"`
	Value     LazyArray  `yaml:"value,omitempty"`
	Order     int        `yaml:"order,omitempty"`
	Key       string     `yaml:"key,omitempty"`
	Required  BoolString `yaml:"required,omitempty"`
	RepeatKey BoolString `yaml:"repeat_key,omitempty"`
	SkipKey   BoolString `yaml:"skip_key,omitempty"`
	Separator string     `yaml:"separator,omitempty"`
}

// UnmarshalYAML populates the instace fields from the
// given data node
func (a *Argument) UnmarshalYAML(value *yaml.Node) error {
	// omit Condition as it can implicitly be set by defining a value
	// a.Condition = trueString
	a.Required = falseString
	a.RepeatKey = trueString
	a.SkipKey = falseString
	a.Separator = " "

	var v string
	if err := value.Decode(&v); err == nil {
		a.Value = []string{v}
		return nil
	}

	type rawArgument Argument
	return value.Decode((*rawArgument)(a))
}

// MarshalIcinga renders the argument in the Icinga config syntax format
func (a *Argument) MarshalIcinga(name string) ([]byte, error) {
	buf := &bytes.Buffer{}

	buf.WriteByte('"')
	buf.WriteString(name)
	buf.WriteString("\" = {\n")

	a.marshalIcingaCondition(buf)
	a.marshalIcingaValue(buf)
	a.marshalIcingaOrder(buf)
	a.marshalIcingaKey(buf)
	a.marshalIcingaRequired(buf)
	a.marshalIcingaRepeatKey(buf)
	a.marshalIcingaSkipKey(buf)
	a.marshalIcingaSeparator(buf)

	buf.WriteString("}\n")

	return buf.Bytes(), nil
}

func (a *Argument) marshalIcingaCondition(buf *bytes.Buffer) {
	if a.Condition != "" {
		b, err := strconv.ParseBool(string(a.Condition))
		if err != nil {
			buf.WriteString(`  set_if = "`)
			buf.WriteString(string(a.Condition))
			buf.WriteByte('"')
		} else if b {
			buf.WriteString(`  set_if = true`)
		} else {
			buf.WriteString(`  set_if = false`)
		}

		buf.WriteByte('\n')
	}
}

func (a *Argument) marshalIcingaValue(buf *bytes.Buffer) {
	if len(a.Value) == 0 {
		return
	}

	if len(a.Value) == 1 {
		buf.WriteString(`  value = "`)
		buf.WriteString(a.Value[0])
		buf.WriteString("\"\n")

		return
	}

	buf.WriteString(`  value = [\n`)
	for _, v := range a.Value {
		buf.WriteString(`    "`)
		buf.WriteString(v)
		buf.WriteString("\",\n")
	}
	buf.WriteString("  ]\n")
}

func (a *Argument) marshalIcingaOrder(buf *bytes.Buffer) {
	if a.Order != 0 {
		fmt.Fprintf(buf, "  order = %d\n", a.Order)
	}
}

func (a *Argument) marshalIcingaKey(buf *bytes.Buffer) {
	if a.Key != "" {
		buf.WriteString(`  key = "`)
		buf.WriteString(string(a.Key))
		buf.WriteString("\"\n")
	}
}

func (a *Argument) marshalIcingaRequired(buf *bytes.Buffer) {
	if a.Required != "" {
		b, err := strconv.ParseBool(string(a.Required))
		if err != nil {
			buf.WriteString(`  required = "`)
			buf.WriteString(string(a.Required))
			buf.WriteByte('"')
		} else if b {
			buf.WriteString(`  required = true`)
		} else {
			buf.WriteString(`  required = false`)
		}

		buf.WriteByte('\n')
	}
}

func (a *Argument) marshalIcingaRepeatKey(buf *bytes.Buffer) {
	if a.RepeatKey != "" {
		b, err := strconv.ParseBool(string(a.RepeatKey))
		if err != nil {
			buf.WriteString(`  repeat_key = "`)
			buf.WriteString(string(a.RepeatKey))
			buf.WriteByte('"')
		} else if b {
			buf.WriteString(`  repeat_key = true`)
		} else {
			buf.WriteString(`  repeat_key = false`)
		}

		buf.WriteByte('\n')
	}
}

func (a *Argument) marshalIcingaSkipKey(buf *bytes.Buffer) {
	if a.SkipKey != "" {
		b, err := strconv.ParseBool(string(a.SkipKey))
		if err != nil {
			buf.WriteString(`  skip_key = "`)
			buf.WriteString(string(a.SkipKey))
			buf.WriteByte('"')
		} else if b {
			buf.WriteString(`  skip_key = true`)
		} else {
			buf.WriteString(`  skip_key = false`)
		}

		buf.WriteByte('\n')
	}
}

func (a *Argument) marshalIcingaSeparator(buf *bytes.Buffer) {
	if a.Separator != "" {
		buf.WriteString(`  separator = "`)
		buf.WriteString(string(a.Separator))
		buf.WriteString("\"\n")
	}
}
