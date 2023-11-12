package nagios

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestJoinKeyValues(t *testing.T) {
	have := map[string]string{
		"":      "value",
		"param": "value",
		"flag":  "",
	}
	got := JoinKeyValues(have, "=")

	assert.Equal(t, 3, len(got))
	assert.Assert(t, cmp.Contains(got, "=value"))
	assert.Assert(t, cmp.Contains(got, "param=value"))
	assert.Assert(t, cmp.Contains(got, "flag="))
}

func TestRenderArguments(t *testing.T) {
	have := []*argument{
		&argument{
			condition: false,
			value:     []string{},
			key:       "--false-condition",
		},
		&argument{
			condition: true,
			value:     []string{"not_used"},
		},
		&argument{
			condition: true,
			value:     []string{},
			key:       "--true-condition",
		},
		&argument{
			condition: true,
			value:     []string{},
			order:     -5,
			key:       "--first-by-order",
		},
		&argument{
			condition: true,
			value:     []string{"", ""},
			key:       "--repeat-empty",
			repeatKey: true,
		},
		&argument{
			condition: true,
			value:     []string{"spec", ""},
			key:       "--optional-value",
			repeatKey: true,
			separator: "=",
		},
		&argument{
			condition: true,
			value:     []string{"one", "two"},
			key:       "--count",
			repeatKey: false,
			separator: " ",
		},
		&argument{
			condition: true,
			value:     []string{"ONE", "TWO"},
			key:       "ARGS",
			skipKey:   true,
			separator: " ",
		},
	}
	want := []string{
		"--first-by-order",
		"--true-condition",
		"--repeat-empty", "--repeat-empty",
		"--optional-value=spec", "--optional-value",
		"--count", "one", "two",
		"ONE", "TWO",
	}
	got := renderArguments(have)

	assert.DeepEqual(t, want, got)
}
