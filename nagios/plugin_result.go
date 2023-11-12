package nagios

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// PluginResult contains the summary of a Nagios plugin execution
type PluginResult struct {
	Status   ExitCode
	Error    error
	Output   string
	Trailer  []string
	PerfData []PerfData
}

// String renders the plugin result in a Nagios compatible way
func (r *PluginResult) String() string {
	prefix := r.Status.String()
	if strings.HasPrefix(r.Output, prefix) {
		prefix = r.Output
	} else {
		prefix = prefix + ": " + r.Output
	}

	if len(r.PerfData) == 0 {
		return prefix
	}

	pd := make([]string, len(r.PerfData))
	for i, p := range r.PerfData {
		pd[i] = p.String()
	}

	return prefix + PerfDataOutputDelimiter + strings.Join(pd, " ")
}

// PluginResultDecoder is a decoder implementation for Nagios plugin output
type PluginResultDecoder struct {
	scanner *bufio.Scanner
}

// NewPluginResultDecoder creates a new decoder instance
// using the give reader as data source
func NewPluginResultDecoder(r io.Reader) *PluginResultDecoder {
	scanner := bufio.NewScanner(r)
	result := &PluginResultDecoder{
		scanner: scanner,
	}

	return result
}

// Decode uses a scanner to drain the internal reader of any data.
// Processed information are fed back into the given result instance.
func (d *PluginResultDecoder) Decode(result *PluginResult) error {
	var trailer bool

	if result.Trailer == nil {
		result.Trailer = []string{}
	}

	if result.PerfData == nil {
		result.PerfData = []PerfData{}
	}

	for d.scanner.Scan() {
		fragments := strings.Split(d.scanner.Text(), PerfDataOutputDelimiter)
		parts := len(fragments)

		if parts == 0 {
			continue
		}

		if parts > 1 {
			p, err := ParsePerfDataOutput(strings.TrimSpace(fragments[1]))
			if err != nil {
				return err
			}

			result.PerfData = append(result.PerfData, p...)
		}

		if parts > 2 {
			return fmt.Errorf("Malformed plugin output with %d perfdata delimiters", parts-1)
		}

		if trailer {
			result.Trailer = append(result.Trailer, strings.TrimSpace(fragments[0]))
		} else {
			result.Output = strings.TrimSpace(fragments[0])
			trailer = true
		}
	}

	if err := d.scanner.Err(); err != nil {
		return err
	}

	return nil
}
