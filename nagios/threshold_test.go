package nagios

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestParseThreshold(t *testing.T) {
	type testCase struct {
		have      string
		wantMiss  []float64
		wantAlert []float64
		wantError bool
	}

	testCases := map[string]testCase{
		"10": testCase{
			wantAlert: []float64{-1.0, 11.0},
			wantMiss:  []float64{0.0, 5.0, 10.0},
			have:      "10",
		},
		"10:": testCase{
			wantAlert: []float64{-1.0, 0.0, 9.0},
			wantMiss:  []float64{10.0, 11.0},
			have:      "10:",
		},
		"~:10": testCase{
			wantAlert: []float64{11.0},
			wantMiss:  []float64{-1.0, 0.0, 9.0, 10.0},
			have:      "~:10",
		},
		"10:20": testCase{
			wantAlert: []float64{-1.0, 9.0, 21.0},
			wantMiss:  []float64{10.0, 15.0, 20.0},
			have:      "10:20",
		},
		"@10:20": testCase{
			wantAlert: []float64{10.0, 11.0, 19.0, 20.0},
			wantMiss:  []float64{-1.0, 0.0, 9.0, 21.0},
			have:      "@10:20",
		},

		"empty": testCase{
			wantAlert: []float64{-1.0},
			wantMiss:  []float64{0.0, 1.0},
			have:      "",
		},
		":": testCase{
			wantAlert: []float64{-1.0},
			wantMiss:  []float64{0.0, 1.0},
			have:      ":",
		},

		"malformed left limit": testCase{
			wantError: true,
			have:      "bac:10",
		},
		"malformed right limit": testCase{
			wantError: true,
			have:      "10:bac",
		},
		"misplaced equality modifier": testCase{
			wantError: true,
			have:      "10:@20",
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			got, err := ParseThreshold(tc.have)

			if tc.wantError {
				assert.Assert(t, err != nil)
				return
			}

			assert.Assert(t, err)

			for _, a := range tc.wantAlert {
				assert.Assert(t, got.Alert(a), "Range definition %s should generate an alert for %f", tc.have, a)
			}

			for _, m := range tc.wantMiss {
				assert.Assert(t, !got.Alert(m), "Range definition %s should not generate an alert for %f", tc.have, m)
			}
		})
	}
}

func TestThresholdEqual(t *testing.T) {
	type testCase struct {
		left  *Threshold
		right *Threshold
		want  bool
	}

	testCases := map[string]testCase{
		"nil": testCase{
			want: true,
		},
		"non-nil": testCase{
			left: NewThreshold(1),
		},

		"outside/inside": testCase{
			left:  NewOutsideThreshold(1, 2),
			right: NewInsideThreshold(1, 2),
		},
		"inside": testCase{
			left:  NewInsideThreshold(1, 2),
			right: NewInsideThreshold(1, 2),
			want:  true,
		},
		"greater/inside": testCase{
			left:  NewGreaterThreshold(1),
			right: NewInsideThreshold(1, 2),
		},
		"lesser/inside": testCase{
			left:  NewLesserThreshold(2),
			right: NewInsideThreshold(1, 2),
		},
		"min/inside": testCase{
			left:  NewThreshold(1),
			right: NewInsideThreshold(1, 2),
		},

		"outside": testCase{
			left:  NewOutsideThreshold(1, 2),
			right: NewOutsideThreshold(1, 2),
			want:  true,
		},
		"inside/outside": testCase{
			left:  NewInsideThreshold(1, 2),
			right: NewOutsideThreshold(1, 2),
		},
		"greater/outside": testCase{
			left:  NewGreaterThreshold(1),
			right: NewOutsideThreshold(1, 2),
		},
		"lesser/outside": testCase{
			left:  NewLesserThreshold(2),
			right: NewOutsideThreshold(1, 2),
		},
		"min/outside": testCase{
			left:  NewThreshold(1),
			right: NewOutsideThreshold(1, 2),
		},

		"outside/greater": testCase{
			left:  NewOutsideThreshold(1, 2),
			right: NewGreaterThreshold(1),
		},
		"inside/greater": testCase{
			left:  NewInsideThreshold(1, 2),
			right: NewGreaterThreshold(1),
		},
		"greater": testCase{
			left:  NewGreaterThreshold(1),
			right: NewGreaterThreshold(1),
			want:  true,
		},
		"lesser/greater": testCase{
			left:  NewLesserThreshold(2),
			right: NewGreaterThreshold(1),
		},
		"min/greater": testCase{
			left:  NewThreshold(1),
			right: NewGreaterThreshold(1),
		},

		"outside/lesser": testCase{
			left:  NewOutsideThreshold(1, 2),
			right: NewLesserThreshold(2),
		},
		"inside/lesser": testCase{
			left:  NewInsideThreshold(1, 2),
			right: NewLesserThreshold(2),
		},
		"greater/lesser": testCase{
			left:  NewGreaterThreshold(1),
			right: NewLesserThreshold(2),
		},
		"lesser": testCase{
			left:  NewLesserThreshold(2),
			right: NewLesserThreshold(2),
			want:  true,
		},
		"min/lesser": testCase{
			left:  NewThreshold(1),
			right: NewLesserThreshold(2),
		},

		"outside/min": testCase{
			left:  NewOutsideThreshold(1, 2),
			right: NewThreshold(1),
		},
		"inside/min": testCase{
			left:  NewInsideThreshold(1, 2),
			right: NewThreshold(1),
		},
		"greater/min": testCase{
			left:  NewGreaterThreshold(1),
			right: NewThreshold(1),
		},
		"lesser/min": testCase{
			left:  NewLesserThreshold(2),
			right: NewThreshold(1),
		},
		"min": testCase{
			left:  NewThreshold(1),
			right: NewThreshold(1),
			want:  true,
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			assert.Assert(t, tc.left.Equal(tc.right) == tc.want, "left=%s; right=%s", tc.left, tc.right)
		})
	}
}

func TestThresholdString(t *testing.T) {
	type testCase struct {
		have *Threshold
		want string
	}

	testCases := map[string]testCase{
		"outside": testCase{
			have: NewOutsideThreshold(10, 20),
			want: "10:20",
		},
		"inside": testCase{
			have: NewInsideThreshold(10, 20),
			want: "@10:20",
		},
		"greater": testCase{
			have: NewGreaterThreshold(10),
			want: "~:10",
		},
		"lesser": testCase{
			have: NewLesserThreshold(10),
			want: "10:",
		},
		"min": testCase{
			have: NewThreshold(10),
			want: "10",
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			got := tc.have.String()

			assert.Equal(t, tc.want, got)
		})
	}
}
