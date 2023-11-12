package nagios

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestParsePerfData(t *testing.T) {
	type testCase struct {
		wantError bool
		want      *PerfData
		have      string
	}

	testCases := map[string]testCase{
		"empty": testCase{
			have:      "",
			wantError: true,
		},
		"empty label": testCase{
			have:      "=1",
			wantError: true,
		},
		"empty quoted label": testCase{
			have:      "''=1",
			wantError: true,
		},
		"empty data": testCase{
			have: "empty=;;;;",
			want: NewUndefinedPerfData("empty"),
		},
		"percent value": testCase{
			have: "pct=50%",
			want: NewValuePerfData("pct", NewPercentValue(50.0)),
		},
		"percent value with limits": testCase{
			want: NewScopedPerfData("limits", NewPercentValue(50.0), 10, 200),
			have: "limits=50%;;;10;200",
		},
		"percent value with thresholds": testCase{
			want: NewThresholdPerfData("ths", NewPercentValue(50.0), NewThreshold(10), NewThreshold(20)),
			have: "ths=50%;10;20",
		},
		"percent value with thresholds and limits": testCase{
			want: NewPerfData("both", NewPercentValue(50.0), NewThreshold(10), NewThreshold(20), 10, 200),
			have: "both=50%;10;20;10;200",
		},
		"unit-less value": testCase{
			want: NewValuePerfData("test", NewFloatValue(123.0)),
			have: "test=123",
		},
		"thresholds": testCase{
			want: NewThresholdPerfData("ths", NewPercentValue(50.0), NewOutsideThreshold(15, 25), NewOutsideThreshold(10, 30)),
			have: "ths=50%;15:25;10:30",
		},
		"undefined value": testCase{
			want: NewUndefinedPerfData("test"),
			have: "test=",
		},
		"quoted label": testCase{
			want: NewValuePerfData("test", NewFloatValue(123.0)),
			have: "'test'=123",
		},
		"label space": testCase{
			want: NewValuePerfData("unit test", NewFloatValue(123.0)),
			have: "'unit test'=123",
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			got, err := ParsePerfData(tc.have)

			if tc.wantError {
				assert.Assert(t, err != nil)
			} else {
				assert.Assert(t, err)
				assert.Assert(t, tc.want.Equal(got), "have=%s; want=%s; got=%s", tc.have, tc.want, got)
			}
		})
	}
}

func TestPerfDataEqual(t *testing.T) {
	type testCase struct {
		left  *PerfData
		right *PerfData
		want  bool
	}

	testCases := map[string]testCase{
		"undefined/undefined": testCase{
			left:  NewUndefinedPerfData("test"),
			right: NewUndefinedPerfData("test"),
			want:  true,
		},
		"undefined/value": testCase{
			left:  NewUndefinedPerfData("test"),
			right: NewValuePerfData("test", NewPercentValue(50)),
		},
		"undefined/threshold": testCase{
			left:  NewUndefinedPerfData("test"),
			right: NewThresholdPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20)),
		},
		"undefined/scoped": testCase{
			left:  NewUndefinedPerfData("test"),
			right: NewScopedPerfData("test", NewPercentValue(50), 10, 200),
		},
		"undefined/custom": testCase{
			left:  NewUndefinedPerfData("test"),
			right: NewPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20), 10, 200),
		},

		"value/undefined": testCase{
			left:  NewValuePerfData("test", NewPercentValue(50)),
			right: NewUndefinedPerfData("test"),
		},
		"value/value": testCase{
			left:  NewValuePerfData("test", NewPercentValue(50)),
			right: NewValuePerfData("test", NewPercentValue(50)),
			want:  true,
		},
		"value/threshold": testCase{
			left:  NewValuePerfData("test", NewPercentValue(50)),
			right: NewThresholdPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20)),
		},
		"value/scoped": testCase{
			left:  NewValuePerfData("test", NewPercentValue(50)),
			right: NewScopedPerfData("test", NewPercentValue(50), 10, 200),
		},
		"value/custom": testCase{
			left:  NewValuePerfData("test", NewPercentValue(50)),
			right: NewPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20), 10, 200),
		},

		"threshold/undefined": testCase{
			left:  NewThresholdPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20)),
			right: NewUndefinedPerfData("test"),
		},
		"threshold/value": testCase{
			left:  NewThresholdPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20)),
			right: NewValuePerfData("test", NewPercentValue(50)),
		},
		"threshold/threshold": testCase{
			left:  NewThresholdPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20)),
			right: NewThresholdPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20)),
			want:  true,
		},
		"threshold/scoped": testCase{
			left:  NewThresholdPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20)),
			right: NewScopedPerfData("test", NewPercentValue(50), 10, 200),
		},
		"threshold/custom": testCase{
			left:  NewThresholdPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20)),
			right: NewPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20), 10, 200),
		},

		"scoped/undefined": testCase{
			left:  NewScopedPerfData("test", NewPercentValue(50), 10, 200),
			right: NewUndefinedPerfData("test"),
		},
		"scoped/value": testCase{
			left:  NewScopedPerfData("test", NewPercentValue(50), 10, 200),
			right: NewValuePerfData("test", NewPercentValue(50)),
		},
		"scoped/threshold": testCase{
			left:  NewScopedPerfData("test", NewPercentValue(50), 10, 200),
			right: NewThresholdPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20)),
		},
		"scoped/scoped": testCase{
			left:  NewScopedPerfData("test", NewPercentValue(50), 10, 200),
			right: NewScopedPerfData("test", NewPercentValue(50), 10, 200),
			want:  true,
		},
		"scoped/custom": testCase{
			left:  NewScopedPerfData("test", NewPercentValue(50), 10, 200),
			right: NewPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20), 10, 200),
		},

		"custom/undefined": testCase{
			left:  NewPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20), 10, 200),
			right: NewUndefinedPerfData("test"),
		},
		"custom/value": testCase{
			left:  NewPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20), 10, 200),
			right: NewValuePerfData("test", NewPercentValue(50)),
		},
		"custom/threshold": testCase{
			left:  NewPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20), 10, 200),
			right: NewThresholdPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20)),
		},
		"custom/scoped": testCase{
			left:  NewPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20), 10, 200),
			right: NewScopedPerfData("test", NewPercentValue(50), 10, 200),
		},
		"custom/custom": testCase{
			left:  NewPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20), 10, 200),
			right: NewPerfData("test", NewPercentValue(50), NewThreshold(10), NewThreshold(20), 10, 200),
			want:  true,
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			assert.Assert(t, tc.left.Equal(tc.right) == tc.want, "left=%s; right=%s", tc.left, tc.right)
		})
	}
}

func TestPerfDataString(t *testing.T) {
	type testCase struct {
		have *PerfData
		want string
	}

	min := 10
	max := 200
	testCases := map[string]testCase{
		"percent value with thresholds and limits": testCase{
			have: &PerfData{
				label: "test",
				value: NewPercentValue(50.0),
				warn:  NewThreshold(10),
				crit:  NewThreshold(20),
				min:   &min,
				max:   &max,
			},
			want: "test=50%;10;20;10;200",
		},
		"percent value with limits": testCase{
			have: &PerfData{
				label: "test",
				value: NewPercentValue(50.0),
				min:   &min,
				max:   &max,
			},
			want: "test=50%;;;10;200",
		},
		"percent value with upper limit": testCase{
			have: &PerfData{
				label: "test",
				value: NewPercentValue(50.0),
				max:   &max,
			},
			want: "test=50%;;;;200",
		},
		"percent value with lower limit": testCase{
			have: &PerfData{
				label: "test",
				value: NewPercentValue(50.0),
				min:   &min,
			},
			want: "test=50%;;;10",
		},
		"percent value with thresholds": testCase{
			have: &PerfData{
				label: "test",
				value: NewPercentValue(50.0),
				warn:  NewThreshold(10),
				crit:  NewThreshold(20),
			},
			want: "test=50%;10;20",
		},
		"percent value with critical threshold": testCase{
			have: &PerfData{
				label: "test",
				value: NewPercentValue(50.0),
				crit:  NewThreshold(20),
			},
			want: "test=50%;;20",
		},
		"percent value with warning threshold": testCase{
			have: &PerfData{
				label: "test",
				value: NewPercentValue(50.0),
				warn:  NewThreshold(10),
			},
			want: "test=50%;10",
		},
		"percent value": testCase{
			have: &PerfData{
				label: "test",
				value: NewPercentValue(50.0),
			},
			want: "test=50%",
		},
		"empty": testCase{
			have: &PerfData{
				label: "test",
			},
			want: "test=U",
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			got := tc.have.String()

			assert.Equal(t, tc.want, got)
		})
	}
}
