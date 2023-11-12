package config

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// LazyArray is an array instance, which can be
// declared using a single-item notation
type LazyArray []string

// String create a newline-delimited string of the instance items
func (s LazyArray) String() string {
	return strings.Join(s, "\n")
}

// UnmarshalYAML populates the instace from the
// given data node
func (s *LazyArray) UnmarshalYAML(value *yaml.Node) error {
	var single string
	if err := value.Decode(&single); err == nil {
		*s = []string{single}
		return nil
	}

	var multi []string
	if err := value.Decode(&multi); err != nil {
		return err
	}

	*s = multi
	return nil
}

// MarshalIcinga renders the array in the Icinga config syntax format
func (s LazyArray) MarshalIcinga(name string) ([]byte, error) {
	buf := &bytes.Buffer{}
	items := make([]string, len(s))

	for i, v := range s {
		items[i] = fmt.Sprintf("%q", v)
	}

	buf.WriteString(name)
	buf.WriteString(" = [")
	buf.WriteString(strings.Join(items, ", "))
	buf.WriteByte(']')

	return buf.Bytes(), nil

}

// NumberDuration is a time.Duration implementation
// which can optionally be declared without any
// time scale (defaulting to seconds)
type NumberDuration time.Duration

// String renders the duration instance in seconds
func (n NumberDuration) String() string {
	d := time.Duration(n)

	return fmt.Sprintf("%.0fs", d.Seconds())
}

// UnmarshalYAML populates the instace from the
// given data node
func (n *NumberDuration) UnmarshalYAML(value *yaml.Node) error {
	var numeric int64
	if err := value.Decode(&numeric); err == nil {
		*n = NumberDuration(numeric * int64(time.Second))
		return nil
	}

	var duration time.Duration
	if err := value.Decode(&duration); err != nil {
		return err
	}

	*n = NumberDuration(duration)
	return nil
}

// BooleanString is a string type, which can be unmarshaled
// from a native bool value
type BoolString string

// UnmarshalYAML populates the instace from the
// given data node
func (b *BoolString) UnmarshalYAML(value *yaml.Node) error {
	var native bool
	if err := value.Decode(&native); err == nil {
		if native {
			*b = "true"
		} else {
			*b = "false"
		}
		return nil
	}

	type rawBoolString BoolString
	return value.Decode((*rawBoolString)(b))
}
