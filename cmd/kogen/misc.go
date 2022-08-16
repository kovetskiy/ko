package main

func push[K any](prefix []K, value K) []K {
	return append(append([]K{}, prefix...), value)
}

func inSlice[K comparable](items []K, target K) bool {
	for _, value := range items {
		if value == target {
			return true
		}
	}
	return false
}
