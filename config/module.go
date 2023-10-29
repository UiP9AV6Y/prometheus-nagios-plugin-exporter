package config

import (
	"bytes"
	"fmt"
	"time"
)

type Module struct {
	Command     []string            `yaml:"command,omitempty"`
	Timeout     time.Duration       `yaml:"timeout,omitempty"`
	Arguments   map[string]Argument `yaml:"arguments,omitempty"`
	Variables   map[string]string   `yaml:"variables,omitempty"`
	Environment map[string]string   `yaml:"environment,omitempty"`
}

func (m *Module) MarshalIcinga(name string) ([]byte, error) {
	buf := &bytes.Buffer{}

	buf.WriteString("object CheckCommand \"")
	buf.WriteString(name)
	buf.WriteString("\" {\n")

	marshalIcingaCommand(buf)
	marshalIcingaTimeout(buf)
	marshalIcingaEnvironment(buf)

	if err := marshalIcingaArguments(buf); err != nil {
		return nil, err
	}

	marshalIcingaVars(buf)

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

		buf.WriteString("    \"")
		buf.WriteString(k)
		buf.WriteString("\" = \"")
		buf.Write(a)
	}

	buf.WriteString("  }\n")
}

func (m *Module) marshalIcingaVars(buf *bytes.Buffer) {
	for k, v := range m.Variables {
		buf.WriteString("  vars.")
		buf.WriteString(k)
		buf.WriteString(" = \"")
		buf.WriteString(v)
		buf.WriteString("\"\n")
	}
}

func (m *Module) marshalIcingaTimeout(buf *bytes.Buffer) {
	buf.WriteString("  timeout = ")
	fmt.Fprintf(buf, "%ds\n", m.Timeout.Seconds())
}

func (m *Module) marshalIcingaCommand(buf *bytes.Buffer) {
	firstCmd := true

	buf.WriteString("  command = [ ")

	for _, cmd := range m.Command {
		if firstCmd {
			firstCmd = false
		} else {
			buf.WriteString(" + ")
		}

		buf.WriteByte('"')
		buf.WriteString(cmd)
		buf.WriteByte('"')
	}

	buf.WriteString(" ]\n")
}
