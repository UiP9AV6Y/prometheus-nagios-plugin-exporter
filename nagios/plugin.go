package nagios

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"strings"
)

// Plugin represents a Nagios plugin execution definition
type Plugin struct {
	command     string
	arguments   []string
	environment []string
}

// NewArgumentPlugin creates a new plugin instance using the given command
// and commandline arguments
func NewArgumentPlugin(command string, arguments ...string) *Plugin {
	return NewPlugin(command, arguments, []string{})
}

// NewPlugin creates a new plugin instance using the given command,
// commandline arguments, and environment variables. The environment
// variables are expected to be an array of key/value pairs, joined by =
func NewPlugin(command string, arguments, environment []string) *Plugin {
	result := &Plugin{
		command:     command,
		arguments:   arguments,
		environment: environment,
	}

	return result
}

// String creates a rudimentary commandline representation,
// using the command and its arguments
func (p *Plugin) String() string {
	s := make([]string, 0, len(p.arguments)+1)
	s = append(s, p.command)
	s = append(s, p.arguments...)

	return strings.Join(s, " ")
}

// Run is a wrapper to exec.Command. Any error is the result
// of the command not being able to be executed. The returned
// PluginResult contains any command output on STDERR as error.
func (p *Plugin) Run(ctx context.Context) (*PluginResult, error) {
	cmd := exec.CommandContext(ctx, p.command, p.arguments...)
	cmd.Env = p.environment

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	result := &PluginResult{}
	if err := NewPluginResultDecoder(stdout).Decode(result); err != nil {
		return nil, err
	}

	errout, err := io.ReadAll(stderr)
	if err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			return nil, err
		}

		result.Status = ExitCode(exitError.ExitCode())
	}

	if len(errout) > 0 {
		result.Error = errors.New(string(errout))
	}

	return result, nil
}
