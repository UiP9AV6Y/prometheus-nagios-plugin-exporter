package config

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"gotest.tools/v3/assert"
)

func TestLazyArrayString(t *testing.T) {
	type testCase struct {
		want string
		have LazyArray
	}

	testCases := map[string]testCase{
		"empty": testCase{
			have: LazyArray([]string{}),
			want: "",
		},
		"single empty": testCase{
			have: LazyArray([]string{""}),
			want: "",
		},
		"two empty": testCase{
			have: LazyArray([]string{"", ""}),
			want: "\n",
		},
		"single value": testCase{
			have: LazyArray([]string{"one"}),
			want: "one",
		},
		"two value": testCase{
			have: LazyArray([]string{"one", "two"}),
			want: "one\ntwo",
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			got := tc.have.String()

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNumberDurationString(t *testing.T) {
	type testCase struct {
		want string
		have NumberDuration
	}

	testCases := map[string]testCase{
		"zero": testCase{
			have: NumberDuration(0 * int64(time.Second)),
			want: "0s",
		},
		"30 seconds": testCase{
			have: NumberDuration(30 * int64(time.Second)),
			want: "30s",
		},
		"60 seconds": testCase{
			have: NumberDuration(60 * int64(time.Second)),
			want: "60s",
		},
		"90 seconds": testCase{
			have: NumberDuration(90 * int64(time.Second)),
			want: "90s",
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			got := tc.have.String()

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestLazyArray(t *testing.T) {
	type testFixture struct {
		Unit LazyArray `yaml:"unit,omitempty"`
	}
	type testCase struct {
		wantError bool
		want      LazyArray
		have      []byte
	}

	testCases := map[string]testCase{
		"nil": testCase{
			have: []byte("unit: ~"),
			want: nil,
		},
		"empty array": testCase{
			have: []byte("unit: []"),
			want: []string{},
		},
		"empty string": testCase{
			have: []byte(`unit: ""`),
			want: []string{""},
		},
		"single item array": testCase{
			have: []byte(`unit: ["item"]`),
			want: []string{"item"},
		},
		"single item string": testCase{
			have: []byte(`unit: "item"`),
			want: []string{"item"},
		},
		"two item array": testCase{
			have: []byte(`unit: ["one", "two"]`),
			want: []string{"one", "two"},
		},
		"multi-line string": testCase{
			have: []byte(`unit: |-
  one
  two
`),
			want: []string{"one\ntwo"},
		},
		"number primitive": testCase{
			have: []byte(`unit: 1`),
			want: []string{"1"},
		},
		"number array": testCase{
			have: []byte(`unit: [1]`),
			want: []string{"1"},
		},
		"bool primitive": testCase{
			have: []byte(`unit: true`),
			want: []string{"true"},
		},
		"bool array": testCase{
			have: []byte(`unit: [true]`),
			want: []string{"true"},
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			var subject testFixture
			err := yaml.Unmarshal(tc.have, &subject)

			if tc.wantError {
				assert.Assert(t, err != nil)
			} else {
				assert.Assert(t, err)
				assert.DeepEqual(t, tc.want, subject.Unit)
			}
		})
	}
}

func TestNumberDuration(t *testing.T) {
	type testFixture struct {
		Unit NumberDuration `yaml:"unit,omitempty"`
	}
	type testCase struct {
		wantError bool
		want      NumberDuration
		have      []byte
	}

	testCases := map[string]testCase{
		"zero": testCase{
			have: []byte("unit: 0"),
		},
		"number": testCase{
			have: []byte("unit: 123"),
			want: NumberDuration(123 * time.Second),
		},
		"seconds": testCase{
			have: []byte("unit: 60s"),
			want: NumberDuration(60 * time.Second),
		},
		"minutes": testCase{
			have: []byte("unit: 1m"),
			want: NumberDuration(60 * time.Second),
		},
		"garbage": testCase{
			have:      []byte("unit: short"),
			wantError: true,
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			var subject testFixture
			err := yaml.Unmarshal(tc.have, &subject)

			if tc.wantError {
				assert.Assert(t, err != nil)
			} else {
				assert.Assert(t, err)
				assert.Equal(t, tc.want, subject.Unit)
			}
		})
	}
}

func TestBoolString(t *testing.T) {
	type testFixture struct {
		Unit BoolString `yaml:"unit,omitempty"`
	}
	type testCase struct {
		wantError bool
		want      BoolString
		have      []byte
	}

	testCases := map[string]testCase{
		"boolean true": testCase{
			have: []byte("unit: true"),
			want: "true",
		},
		"boolean false": testCase{
			have: []byte("unit: false"),
			want: "false",
		},
		"numeric true": testCase{
			have: []byte("unit: 1"),
			want: "1",
		},
		"numeric false": testCase{
			have: []byte("unit: 0"),
			want: "0",
		},
		"string true": testCase{
			have: []byte(`unit: "true"`),
			want: "true",
		},
		"string false": testCase{
			have: []byte(`unit: "false"`),
			want: "false",
		},
		"string yes": testCase{
			have: []byte(`unit: "yes"`),
			want: "true",
		},
		"string no": testCase{
			have: []byte(`unit: "no"`),
			want: "false",
		},
		"empty string": testCase{
			have: []byte(`unit: ""`),
			want: "",
		},
		"garbage": testCase{
			have: []byte(`unit: "short"`),
			want: "short",
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			var subject testFixture
			err := yaml.Unmarshal(tc.have, &subject)

			if tc.wantError {
				assert.Assert(t, err != nil)
			} else {
				assert.Assert(t, err)
				assert.Equal(t, tc.want, subject.Unit)
			}
		})
	}
}
