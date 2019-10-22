package util

func StringsMatch(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func UniqueStringSlice(slice []string) []string {
	sMap := make(map[string]bool)
	var dd []string

	for _, k := range slice {
		if _, v := sMap[k]; !v {
			sMap[k] = true
			dd = append(dd, k)
		}
	}

	return dd
}
