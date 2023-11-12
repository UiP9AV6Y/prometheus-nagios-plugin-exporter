package nagios

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestCompact(t *testing.T) {
	type testCase struct {
		have []string
		want []string
	}

	testCases := map[string]testCase{
		"nil": testCase{},
		"empty": testCase{
			have: []string{},
			want: []string{},
		},
		"empty item": testCase{
			have: []string{""},
			want: []string{},
		},
		"empty items": testCase{
			have: []string{"", "", ""},
			want: []string{},
		},
		"empty center": testCase{
			have: []string{"one", "", "two"},
			want: []string{"one", "two"},
		},
		"empty left": testCase{
			have: []string{"", "one", "two"},
			want: []string{"one", "two"},
		},
		"empty right": testCase{
			have: []string{"one", "two", ""},
			want: []string{"one", "two"},
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			got := Compact(tc.have)

			assert.DeepEqual(t, tc.want, got)
		})
	}
}
