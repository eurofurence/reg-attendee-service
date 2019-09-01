package config

import (
	"sort"
)

func sortedKeys(choiceMap *map[string]choiceConfig) []string {
	keys := make([]string, len(*choiceMap))
	i := 0
	for k := range *choiceMap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
