package config

import (
	"bytes"
	"context"
	"fmt"
)

// Module defines a reusable monitoring execution plan
type Module struct {
	Command     string               `yaml:"command,omitempty"`
	Timeout     NumberDuration       `yaml:"timeout,omitempty"`
	Arguments   map[string]Argument  `yaml:"arguments,omitempty"`
	Variables   map[string]LazyArray `yaml:"variables,omitempty"`
	Environment map[string]string    `yaml:"environment,omitempty"`
}

type contextKey string

var (
	nameKey = contextKey("moduleName")
	dataKey = contextKey("moduleData")
)

// NewContext returns a new Context containing the given module and name.
func NewContext(ctx context.Context, name string, module *Module) context.Context {
	ctx = context.WithValue(ctx, nameKey, name)
	return context.WithValue(ctx, dataKey, module)
}

// FromContext returns the Module (+ name) value stored in ctx, if any.
func FromContext(ctx context.Context) (string, *Module, bool) {
	name, n := ctx.Value(nameKey).(string)
	module, m := ctx.Value(dataKey).(*Module)
	return name, module, n && m
}

// MarshalIcinga renders the module in the Icinga config syntax format
func (m *Module) MarshalIcinga(name string) ([]byte, error) {
	buf := &bytes.Buffer{}

	buf.WriteString("object CheckCommand \"")
	buf.WriteString(name)
	buf.WriteString("\" {\n")

	m.marshalIcingaCommand(buf)
	m.marshalIcingaTimeout(buf)
	m.marshalIcingaEnvironment(buf)

	if err := m.marshalIcingaArguments(buf); err != nil {
		return nil, err
	}

	if err := m.marshalIcingaVars(buf); err != nil {
		return nil, err
	}

	buf.WriteByte('}')

	return buf.Bytes(), nil
}

func (m *Module) marshalIcingaEnvironment(buf *bytes.Buffer) {
	buf.WriteString("  env = {\n")

	for k, v := range m.Environment {
		buf.WriteString("    \"")
		buf.WriteString(k)
		buf.WriteString("\" = \"")
		buf.WriteString(v)
		buf.WriteString("\"\n")
	}

	buf.WriteString("  }\n")
}

func (m *Module) marshalIcingaArguments(buf *bytes.Buffer) error {
	buf.WriteString("  arguments = {\n")

	for k, v := range m.Arguments {
		a, err := v.MarshalIcinga(k)
		if err != nil {
			return err
		}

		buf.WriteString("    ")
		buf.Write(a)
	}

	buf.WriteString("  }\n")

	return nil
}

func (m *Module) marshalIcingaVars(buf *bytes.Buffer) error {
	for k, v := range m.Variables {
		i, err := v.MarshalIcinga(k)
		if err != nil {
			return err
		}

		buf.WriteString("  vars.")
		buf.Write(i)
		buf.WriteString("\n")
	}

	return nil
}

func (m *Module) marshalIcingaTimeout(buf *bytes.Buffer) {
	buf.WriteString("  timeout = ")
	fmt.Fprintf(buf, "%s\n", m.Timeout)
}

func (m *Module) marshalIcingaCommand(buf *bytes.Buffer) {
	buf.WriteString("  command = [ ")
	fmt.Fprintf(buf, "%q", m.Command)
	buf.WriteString(" ]\n")
}
