package parser

import "sort"

// Uniq removes duplicates from slice
func Uniq(slice []string) []string {
	uniq := map[string]struct{}{}
	for _, k := range slice {
		uniq[k] = struct{}{}
	}

	keys := MapKeys(uniq)
	sort.Strings(keys)
	return keys
}

// MapKeys returns map keys only
func MapKeys[T comparable, V any](data map[T]V) []T {
	keys := make([]T, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}
