package nagios

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	ThresholdDelimiter = ":"
)

// Alert is a comparison functor for the given value
type Alert func(float64) bool

// Threshold is a boundary definition for peformance metrics
type Threshold struct {
	leftLimit, rightLimit string
	leftAlert, rightAlert Alert
	cmpAnd                bool
}

// LessThan create an alert instance acting on the given value as boundary.
// Any compared value must be less than the comparison in order to pass
func LessThan(cmp float64) Alert {
	return func(value float64) bool {
		return value < cmp
	}
}

// GreaterThan create an alert instance acting on the given value as boundary.
// Any compared value must be greater than the comparison in order to pass
func GreaterThan(cmp float64) Alert {
	return func(value float64) bool {
		return value > cmp
	}
}

// LessEqualThan create an alert instance acting on the given value as boundary.
// Any compared value must be less than or equal to the comparison in order to pass
func LessEqualThan(cmp float64) Alert {
	return func(value float64) bool {
		return value <= cmp
	}
}

// GreaterEqualThan create an alert instance acting on the given value as boundary.
// Any compared value must be greater than or equal to the comparison in order to pass
func GreaterEqualThan(cmp float64) Alert {
	return func(value float64) bool {
		return value >= cmp
	}
}

// True create an alert instance which always passes
func True() Alert {
	return func(_ float64) bool {
		return true
	}
}

// True create an alert instance which always fails
func False() Alert {
	return func(_ float64) bool {
		return false
	}
}

// NewOutsideThreshold creates a threshold instance, alerting on any
// metric outside of the given boundaries
func NewOutsideThreshold(lowerLimit, upperLimit float64) *Threshold {
	l := strconv.FormatFloat(lowerLimit, 'g', -1, 64)
	r := strconv.FormatFloat(upperLimit, 'g', -1, 64)
	result := &Threshold{
		leftLimit:  l,
		rightLimit: r,
		leftAlert:  LessThan(lowerLimit),
		rightAlert: GreaterThan(upperLimit),
	}

	return result
}

// NewInsideThreshold creates a threshold instance, alerting on any
// metric inside of the given boundaries
func NewInsideThreshold(lowerLimit, upperLimit float64) *Threshold {
	l := strconv.FormatFloat(lowerLimit, 'g', -1, 64)
	r := strconv.FormatFloat(upperLimit, 'g', -1, 64)
	result := &Threshold{
		leftLimit:  "@" + l,
		rightLimit: r,
		leftAlert:  GreaterEqualThan(lowerLimit),
		rightAlert: LessEqualThan(upperLimit),
	}

	return result
}

// NewGreaterThreshold creates a threshold instance, alerting on any
// metric greater then the given boundary
func NewGreaterThreshold(minValue float64) *Threshold {
	r := strconv.FormatFloat(minValue, 'g', -1, 64)
	result := &Threshold{
		leftLimit:  "~",
		rightLimit: r,
		leftAlert:  False(),
		rightAlert: GreaterThan(minValue),
	}

	return result
}

// NewLesserThreshold creates a threshold instance, alerting on any
// metric less then the given boundary
func NewLesserThreshold(maxValue float64) *Threshold {
	l := strconv.FormatFloat(maxValue, 'g', -1, 64)
	result := &Threshold{
		leftLimit:  l,
		rightLimit: "",
		leftAlert:  LessThan(maxValue),
		rightAlert: False(),
	}

	return result
}

// NewThreshold creates a threshold instance, alerting on any
// metric less than zero or greater then the given boundary
func NewThreshold(minValue float64) *Threshold {
	r := strconv.FormatFloat(minValue, 'g', -1, 64)
	result := &Threshold{
		leftLimit:  "",
		rightLimit: r,
		leftAlert:  LessThan(0),
		rightAlert: GreaterThan(minValue),
	}

	return result
}

// ParseThreshold parses the given value for alert boundaries
func ParseThreshold(s string) (*Threshold, error) {
	var left, right string
	fragments := strings.Split(s, ThresholdDelimiter)
	parts := len(fragments)

	if parts > 2 {
		return nil, fmt.Errorf("Malformed threshold limits")
	}

	if parts == 1 {
		right = fragments[0]
	} else if parts == 2 {
		left = fragments[0]
		right = fragments[1]
	}

	l, eql, err := parseLeftAlert(left)
	if err != nil {
		return nil, err
	}

	r, err := parseRightAlert(right, eql)
	if err != nil {
		return nil, err
	}

	result := &Threshold{
		leftLimit:  left,
		rightLimit: right,
		leftAlert:  l,
		rightAlert: r,
		cmpAnd:     eql,
	}

	return result, nil
}

func parseLeftAlert(s string) (alert Alert, eql bool, err error) {
	var limit float64

	if s == "" {
		alert = LessThan(0)

		return
	}

	if s == "~" {
		alert = False()

		return
	}

	if s == "@" {
		err = fmt.Errorf("Missing limit in lower threshold value")

		return
	}

	if s[0] == '@' {
		limit, err = strconv.ParseFloat(s[1:], 64)
		eql = true
	} else {
		limit, err = strconv.ParseFloat(s, 64)
	}
	if err != nil {
		return
	}

	if eql {
		alert = GreaterEqualThan(limit)

		return
	}

	alert = LessThan(limit)

	return
}

func parseRightAlert(s string, eql bool) (alert Alert, err error) {
	var limit float64

	if s == "" || s == "~" {
		alert = False()

		return
	}

	limit, err = strconv.ParseFloat(s, 64)
	if err != nil {
		return
	}

	if eql {
		alert = LessEqualThan(limit)

		return
	}

	alert = GreaterThan(limit)

	return
}

// Alert compares the given value against the internal boundaries
func (t *Threshold) Alert(value float64) bool {
	if t.cmpAnd {
		return t.leftAlert(value) && t.rightAlert(value)
	}

	return t.leftAlert(value) || t.rightAlert(value)
}

// String renders the threshold in a Nagios compatible format
func (t *Threshold) String() string {
	if t.leftLimit == "" {
		return t.rightLimit
	}

	return t.leftLimit + ":" + t.rightLimit
}

// Equal comparse the internal fields with those of o
func (t *Threshold) Equal(o *Threshold) bool {
	if o == nil {
		return t == nil
	}

	return t.leftLimit == o.leftLimit && t.rightLimit == o.rightLimit && t.cmpAnd == o.cmpAnd
}
