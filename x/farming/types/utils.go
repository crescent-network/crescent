package types

import "time"

// ParseTime parses string time to time in RFC3339 format.
// This is used only for internal testing purpose.
func ParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
