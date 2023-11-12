package nagios

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func envProviderMock(s string) string {
	if strings.HasPrefix(s, "MATCH_") {
		return s[6:]
	}

	return ""
}

func varsProviderMock(s string) []string {
	if strings.HasPrefix(s, "match_") {
		return []string{s[6:]}
	}

	return nil
}

func TestNewBuilderContextEnv(t *testing.T) {
	type testCase struct {
		have      map[string]string
		want      map[string]string
		wantError bool
	}

	testCases := map[string]testCase{
		"empty": testCase{
			have: map[string]string{},
			want: map[string]string{},
		},
		"miss without default": testCase{
			have: map[string]string{
				"MISS_test": "",
			},
			want: map[string]string{},
		},
		"provider without default": testCase{
			have: map[string]string{
				"MATCH_test": "",
			},
			want: map[string]string{
				"MATCH_test": "test",
			},
		},
		"provider with default": testCase{
			have: map[string]string{
				"MATCH_test": "not used",
			},
			want: map[string]string{
				"MATCH_test": "test",
			},
		},
		"miss with default": testCase{
			have: map[string]string{
				"MISS_test": "fallback",
			},
			want: map[string]string{
				"MISS_test": "fallback",
			},
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			subject := &PluginBuilderContext{
				Env: tc.have,
			}
			subject.VisitEnvironment(envProviderMock)
			assert.DeepEqual(t, tc.want, subject.Env)
		})
	}
}

func TestPluginBuilderContextVisitVariables(t *testing.T) {
	type testCase struct {
		have map[string][]string
		want map[string][]string
	}

	testCases := map[string]testCase{
		"empty": testCase{
			have: map[string][]string{},
			want: map[string][]string{},
		},
		"miss without default": testCase{
			have: map[string][]string{
				"miss_test": []string{""},
			},
			want: map[string][]string{},
		},
		"match without default": testCase{
			have: map[string][]string{
				"match_test": []string{""},
			},
			want: map[string][]string{
				"match_test": []string{"test"},
			},
		},
		"match with default": testCase{
			have: map[string][]string{
				"match_test": []string{"not used"},
			},
			want: map[string][]string{
				"match_test": []string{"test"},
			},
		},
		"miss with default": testCase{
			have: map[string][]string{
				"miss_test": []string{"fallback"},
			},
			want: map[string][]string{
				"miss_test": []string{"fallback"},
			},
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			subject := &PluginBuilderContext{
				Vars: tc.have,
			}
			subject.VisitVariables(varsProviderMock)
			assert.DeepEqual(t, tc.want, subject.Vars)
		})
	}
}
