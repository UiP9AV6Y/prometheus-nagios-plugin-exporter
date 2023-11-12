package template

import (
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	// error message from net/ipsock.go when
	// providing net.SplitHostPort with a value
	// without a port
	netMissingPort = "missing port in address"
)

// NetHost returns the host result from net.SplitHostPort.
// Any error is discarded and an empty string is returned instead
func NetHost(hostport string) string {
	host, _, err := net.SplitHostPort(hostport)
	if err == nil {
		return host
	}

	aerr, ok := err.(*net.AddrError)
	if ok && aerr.Err == netMissingPort {
		// clean up IPv6 notation manually
		return strings.Trim(hostport, "[]")
	}

	return ""
}

// NetPort returns the port result from net.SplitHostPort.
// The value is additionally parsed and transformed into
// a numeric value.
// Any error is discarded and zero is returned instead
func NetPort(hostport string) int {
	_, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return 0
	}

	i, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		return 0
	}

	return int(i)
}

// ReadFile returns the content of the given file as string.
// Any error is discarded and an empty string is returned instead
func ReadFile(f string) string {
	b, err := os.ReadFile(f)
	if err != nil {
		return ""
	}

	return string(b)
}

// Lines joins the given slice using a newline as delimiter
func Lines(s []string) string {
	return strings.Join(s, "\n")
}
