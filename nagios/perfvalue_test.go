package nagios

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestParsePerfValue(t *testing.T) {
	type testCase struct {
		have      string
		want      *PerfValue
		wantError bool
	}

	testCases := map[string]testCase{
		"empty": testCase{
			have: "",
			want: &PerfValue{
				Undef: true,
			},
		},
		"undef": testCase{
			have: "U",
			want: &PerfValue{
				Undef: true,
			},
		},
		"undef alt": testCase{
			have: "u",
			want: &PerfValue{
				Undef: true,
			},
		},
		"integer": testCase{
			have: "123",
			want: &PerfValue{
				Value: 123.0,
			},
		},
		"float": testCase{
			have: "12.3",
			want: &PerfValue{
				Value: 12.3,
			},
		},
		"percent": testCase{
			have: "25%",
			want: &PerfValue{
				Value: 25,
				Unit:  "%",
			},
		},
		"rpm": testCase{
			have: "9001rpm",
			want: &PerfValue{
				Value: 9001,
				Unit:  "rpm",
			},
		},
		"no value": testCase{
			have: "rpm",
			want: &PerfValue{
				Unit: "rpm",
			},
		},
		"abbrev float": testCase{
			have: ".5ppm",
			want: &PerfValue{
				Value: 0.5,
				Unit:  "ppm",
			},
		},
		"malformed float": testCase{
			have:      "1.2.5",
			wantError: true,
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			got, err := ParsePerfValue(tc.have)

			if tc.wantError {
				assert.Assert(t, err != nil)
			} else {
				assert.Assert(t, err)
				assert.Assert(t, cmp.DeepEqual(tc.want, got), "PerfValue(%s)", tc.have)
			}
		})
	}
}

func TestPerfValueEqual(t *testing.T) {
	type testCase struct {
		left  *PerfValue
		right *PerfValue
		want  bool
	}

	testCases := map[string]testCase{
		"nil": testCase{
			want: true,
		},
		"non-nil": testCase{
			left: &PerfValue{},
		},
		"empty": testCase{
			left:  &PerfValue{},
			right: &PerfValue{},
			want:  true,
		},
		"undef": testCase{
			left: &PerfValue{
				Undef: true,
			},
			right: &PerfValue{
				Undef: true,
			},
			want: true,
		},
		"value": testCase{
			left: &PerfValue{
				Value: 1.2,
			},
			right: &PerfValue{
				Value: 1.2,
			},
			want: true,
		},
		"unit": testCase{
			left: &PerfValue{
				Value: 1.2,
				Unit:  "test",
			},
			right: &PerfValue{
				Value: 1.2,
				Unit:  "test",
			},
			want: true,
		},
		"undef superfluous": testCase{
			left: &PerfValue{
				Undef: true,
			},
			right: &PerfValue{
				Undef: true,
				Unit:  "test",
			},
		},
		"undef value": testCase{
			left: &PerfValue{
				Undef: true,
			},
			right: &PerfValue{
				Value: 1.2,
			},
		},
		"unit less": testCase{
			left: &PerfValue{
				Value: 1.2,
				Unit:  "test",
			},
			right: &PerfValue{
				Value: 1.2,
			},
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			assert.Assert(t, tc.left.Equal(tc.right) == tc.want, "left=%s; right=%s", tc.left, tc.right)
		})
	}
}

func TestPerfValueString(t *testing.T) {
	type testCase struct {
		have *PerfValue
		want string
	}

	testCases := map[string]testCase{
		"empty": testCase{
			have: &PerfValue{},
			want: "0",
		},
		"undef": testCase{
			have: NewUndefinedValue(),
			want: "U",
		},
		"float": testCase{
			have: NewFloatValue(1.203),
			want: "1.203",
		},
		"percent": testCase{
			have: NewPercentValue(95.0001),
			want: "95.0001%",
		},
		"rpm": testCase{
			have: NewUnitValue(9001, "rpm"),
			want: "9001rpm",
		},
		"dot zero one": testCase{
			have: NewFloatValue(0.01),
			want: "0.01",
		},
		"dot zero": testCase{
			have: NewFloatValue(0.0),
			want: "0",
		},
		"trailing zeroes": testCase{
			have: NewFloatValue(500.0),
			want: "500",
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			got := tc.have.String()

			assert.Equal(t, tc.want, got)
		})
	}
}
