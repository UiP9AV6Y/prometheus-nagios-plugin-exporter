package nagios

//go:generate stringer -type=ExitCode

// Nagios plugin exit state
type ExitCode int

const (
	OK ExitCode = iota
	WARNING
	CRITICAL
	UNKNOWN
	DEPENDENT
)
