package nagios

import (
	"strconv"
)

// PerfValue represents a performance metric value
type PerfValue struct {
	Value float64
	Undef bool
	Unit  string
}

// NewFloatValue creates a new numeric performance value
// without any unit of measurement
func NewFloatValue(f float64) *PerfValue {
	return NewUnitValue(f, "")
}

// NewPercentValue creates a new numeric peformance value
// using percent as the unit of measurement
func NewPercentValue(f float64) *PerfValue {
	return NewUnitValue(f, "%")
}

// NewUnitValue creates a new numeric peformance value
// using the provided unit of measurement
func NewUnitValue(f float64, u string) *PerfValue {
	result := &PerfValue{
		Value: f,
		Unit:  u,
	}

	return result
}

// NewUndefinedValue creates a new undefined peformance value
func NewUndefinedValue() *PerfValue {
	result := &PerfValue{
		Undef: true,
	}

	return result
}

// ParsePerfValue parses the given string for a performance metric value
func ParsePerfValue(s string) (*PerfValue, error) {
	var i int
	var err error
	result := &PerfValue{}

	if s == "U" || s == "u" || s == "" {
		result.Undef = true
		return result, nil
	}

	for ; i < len(s); i++ {
		if !('0' <= s[i] && s[i] <= '9' || s[i] == '.') {
			break
		}
	}

	if i > 0 {
		result.Value, err = strconv.ParseFloat(s[0:i], 64)
		if err != nil {
			return nil, err
		}
	}

	if i < len(s) {
		result.Unit = s[i:]
	}

	return result, nil
}

// String renders the performance value according to its internal representation.
// An undefined value simple yields U, otherwise the numeric value and optional
// unit of measurement are concatenated and returned
func (v *PerfValue) String() string {
	if v.Undef {
		return "U"
	}

	return strconv.FormatFloat(v.Value, 'g', -1, 64) + v.Unit
}

// Equal comparse the internal fields with those of o
func (v *PerfValue) Equal(o *PerfValue) bool {
	if o == nil {
		return v == nil
	}

	return v.Value == o.Value && v.Undef == o.Undef && v.Unit == o.Unit
}
