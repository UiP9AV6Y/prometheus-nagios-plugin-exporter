package nagios

// Compact removes any empty string from the provided array
func Compact(s []string) []string {
	if s == nil {
		return nil
	}

	result := make([]string, 0, len(s))
	for _, i := range s {
		if i != "" {
			result = append(result, i)
		}
	}

	return result
}
