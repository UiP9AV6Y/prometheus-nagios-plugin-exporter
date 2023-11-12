package template

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestNetHost(t *testing.T) {
	type testCase struct {
		have string
		want string
	}

	testCases := map[string]testCase{
		"IPv6 without port": testCase{
			have: "[::1]",
			want: "::1",
		},
		"IPv6 with port": testCase{
			have: "[::1]:123",
			want: "::1",
		},
		"IPv4 without port": testCase{
			have: "127.1.2.3",
			want: "127.1.2.3",
		},
		"IPv4 with port": testCase{
			have: "127.1.2.3:123",
			want: "127.1.2.3",
		},
		"hostname without port": testCase{
			have: "localhost",
			want: "localhost",
		},
		"hostname with port": testCase{
			have: "localhost:123",
			want: "localhost",
		},
		"fqdn without port": testCase{
			have: "localhost.localdomain",
			want: "localhost.localdomain",
		},
		"fqdn with port": testCase{
			have: "localhost.localdomain:123",
			want: "localhost.localdomain",
		},
		"empty without port": testCase{
			have: "",
			want: "",
		},
		"empty with port": testCase{
			have: ":123",
			want: "",
		},
		"malformed ipv6": testCase{
			have: "::1",
			want: "",
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			got := NetHost(tc.have)

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNetPort(t *testing.T) {
	type testCase struct {
		have string
		want int
	}

	testCases := map[string]testCase{
		"IPv6 without port": testCase{
			have: "[::1]",
			want: 0,
		},
		"IPv6 with port": testCase{
			have: "[::1]:123",
			want: 123,
		},
		"IPv4 without port": testCase{
			have: "127.1.2.3",
			want: 0,
		},
		"IPv4 with port": testCase{
			have: "127.1.2.3:123",
			want: 123,
		},
		"hostname without port": testCase{
			have: "localhost",
			want: 0,
		},
		"hostname with port": testCase{
			have: "localhost:123",
			want: 123,
		},
		"fqdn without port": testCase{
			have: "localhost.localdomain",
			want: 0,
		},
		"fqdn with port": testCase{
			have: "localhost.localdomain:123",
			want: 123,
		},
		"empty without port": testCase{
			have: "",
			want: 0,
		},
		"empty with port": testCase{
			have: ":123",
			want: 123,
		},
		"malformed ipv6": testCase{
			have: "::1",
			want: 0,
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			got := NetPort(tc.have)

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestReadFile(t *testing.T) {
	type testCase struct {
		have string
		want string
	}

	testCases := map[string]testCase{
		"empty result": testCase{
			have: "testdata/not_found",
			want: "",
		},
		"secret": testCase{
			have: "testdata/secret",
			want: "test\n",
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			got := ReadFile(tc.have)

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestLines(t *testing.T) {
	type testCase struct {
		have []string
		want string
	}

	testCases := map[string]testCase{
		"empty": testCase{
			have: []string{},
			want: "",
		},
		"single": testCase{
			have: []string{"one"},
			want: "one",
		},
		"duo": testCase{
			have: []string{"one", "two"},
			want: "one\ntwo",
		},
	}

	for ctx, tc := range testCases {
		t.Run(ctx, func(t *testing.T) {
			got := Lines(tc.have)

			assert.Equal(t, tc.want, got)
		})
	}
}
