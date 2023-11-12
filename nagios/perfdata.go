package nagios

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	PerfDataOutputDelimiter = "|"
	PerfDataLabelDelimiter  = "="
	PerfDataValueDelimiter  = ";"
)

// PerfData holds a single performand data metric and its context (thresholds, limits, ...)
type PerfData struct {
	label string
	value *PerfValue
	warn  *Threshold
	crit  *Threshold
	min   *int
	max   *int
}

// NewUndefinedPerfData creates a new instance with the semantic of the value being undefined
func NewUndefinedPerfData(label string) *PerfData {
	return NewValuePerfData(label, NewUndefinedValue())
}

// NewValuePerfData creates a new instance with the given performance metric
func NewValuePerfData(label string, value *PerfValue) *PerfData {
	return NewThresholdPerfData(label, value, nil, nil)
}

// NewThresholdPerfData creates a new instance with the given performance metric and thresholds
func NewThresholdPerfData(label string, value *PerfValue, warn, crit *Threshold) *PerfData {
	result := &PerfData{
		label: label,
		value: value,
		warn:  warn,
		crit:  crit,
	}

	return result
}

// NewScopedPerfData creates a new instance with the given performance metric and limits
func NewScopedPerfData(label string, value *PerfValue, min, max int) *PerfData {
	return NewPerfData(label, value, nil, nil, min, max)
}

// NewPerfData creates a new instance with the given performance metric, thresholds, and limits
func NewPerfData(label string, value *PerfValue, warn, crit *Threshold, min, max int) *PerfData {
	result := &PerfData{
		label: label,
		value: value,
		warn:  warn,
		crit:  crit,
		min:   &min,
		max:   &max,
	}

	return result
}

// ParsePerfDataOutput parses the given string for performance metrics
// in the Nagios PerfData format
func ParsePerfDataOutput(s string) ([]PerfData, error) {
	fields := strings.Fields(s)
	result := make([]PerfData, len(fields))

	for i, data := range fields {
		perfdata, err := ParsePerfData(data)
		if err != nil {
			return nil, err
		}

		result[i] = *perfdata
	}

	return result, nil
}

// ParsePerfDataOutput parses the given string for a single
// performance metric in the Nagios PerfData format
func ParsePerfData(s string) (*PerfData, error) {
	fragments := strings.Split(s, PerfDataLabelDelimiter)
	parts := len(fragments)
	var label string

	if parts > 0 {
		label = strings.Trim(fragments[0], "'")
	}

	if label == "" {
		return nil, fmt.Errorf("Performance data label must not be empty")
	}

	if parts > 2 {
		return nil, fmt.Errorf("Malformed performance data with too many (%d) labels", parts-1)
	}

	if parts == 1 {
		return NewUndefinedPerfData(label), nil
	}

	result := NewUndefinedPerfData(label)
	err := parsePerfDataValues(result, fragments[1])
	if err != nil {
		return nil, err
	}

	return result, nil
}

func parsePerfDataValues(result *PerfData, s string) (err error) {
	fragments := strings.Split(s, PerfDataValueDelimiter)
	parts := len(fragments)

	if parts > 0 {
		result.value, err = ParsePerfValue(fragments[0])
		if err != nil {
			return err
		}
	}

	if parts > 1 && fragments[1] != "" {
		result.warn, err = ParseThreshold(fragments[1])
		if err != nil {
			return err
		}
	}

	if parts > 2 && fragments[2] != "" {
		result.crit, err = ParseThreshold(fragments[2])
		if err != nil {
			return err
		}
	}

	if parts > 3 && fragments[3] != "" {
		min, err := strconv.Atoi(fragments[3])
		if err != nil {
			return err
		}

		result.min = &min
	}

	if parts > 4 && fragments[4] != "" {
		max, err := strconv.Atoi(fragments[4])
		if err != nil {
			return err
		}

		result.max = &max
	}

	return nil
}

// Label returns the peformance data label
func (d *PerfData) Label() string {
	return d.label
}

// QuotedLabel returns the peformance data label;
// quoted if it contains any spaces
func (d *PerfData) QuotedLabel() string {
	if strings.ContainsRune(d.label, 32) {
		return "'" + d.label + "'"
	}

	return d.label
}

// Min returns the lower peformance data limit
func (d *PerfData) Min() (result int) {
	if d.min != nil {
		result = *d.min
	}

	return
}

// Max returns the upper peformance data limit
func (d *PerfData) Max() (result int) {
	if d.max != nil {
		if *d.max == 0 && d.value != nil && d.value.Unit == "%" {
			result = 100
		} else {
			result = *d.max
		}
	}

	return
}

// Value returns the current peformance data value, or U
// if no such information is available
func (d *PerfData) Value() string {
	if d.value == nil {
		return "U"
	}

	return d.value.String()
}

// Float parses the Value() for numeric data or returns 0 otherwise.
func (d *PerfData) Float() float64 {
	if d.value == nil {
		return 0.0
	}

	return d.value.Value
}

// Warning returns the warning threshold
func (d *PerfData) Warning() string {
	if d.warn == nil {
		return ""
	}

	return d.warn.String()
}

// Warning returns the critical threshold
func (d *PerfData) Critical() string {
	if d.crit == nil {
		return ""
	}

	return d.crit.String()
}

// WarningAlert compares the value against the warning threshold.
// If either of those is not available, the function returns false.
func (d *PerfData) WarningAlert() bool {
	if d.warn == nil || d.value == nil || d.value.Undef {
		return false
	}

	return !d.warn.Alert(d.value.Value)
}

// CriticalAlert compares the value against the critical threshold.
// If either of those is not available, the function returns false.
func (d *PerfData) CriticalAlert() bool {
	if d.crit == nil || d.value == nil || d.value.Undef {
		return false
	}

	return !d.crit.Alert(d.value.Value)
}

// String formats the internal data using the Nagios performance data notation
func (d *PerfData) String() string {
	params := make([]string, 5)

	params[0] = d.Value()
	params[1] = d.Warning()
	params[2] = d.Critical()

	if d.min != nil {
		params[3] = strconv.Itoa(*d.min)
	} else if d.max != nil {
		params[3] = ""
	}

	if d.max != nil {
		params[4] = strconv.Itoa(*d.max)
	}

	for {
		i := len(params) - 1
		if i < 0 || params[i] != "" {
			break
		}

		params = params[:i]
	}

	return d.QuotedLabel() + PerfDataLabelDelimiter + strings.Join(params, PerfDataValueDelimiter)
}

// Equal comparse the internal fields with those of o
func (d *PerfData) Equal(o *PerfData) bool {
	if o == nil {
		return d == nil
	}

	if d.label != o.label {
		return false
	}

	if d.value != nil {
		if !d.value.Equal(o.value) {
			return false
		}
	} else if o.value != nil {
		return false
	}

	if d.warn != nil {
		if !d.warn.Equal(o.warn) {
			return false
		}
	} else if o.warn != nil {
		return false
	}

	if d.crit != nil {
		if !d.crit.Equal(o.crit) {
			return false
		}
	} else if o.crit != nil {
		return false
	}

	if d.min != nil {
		if o.min == nil || *d.min != *o.min {
			return false
		}
	} else if o.min != nil {
		return false
	}

	if d.max != nil {
		if o.max == nil || *d.max != *o.max {
			return false
		}
	} else if o.max != nil {
		return false
	}

	return true
}
