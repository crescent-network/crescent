package cli

// excConditions returns true when exactly one condition is true.
func excConditions(conditions ...bool) bool {
	cnt := 0
	for _, condition := range conditions {
		if condition {
			cnt++
		}
	}
	return cnt == 1
}
