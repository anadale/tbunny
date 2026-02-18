package utils

import "sort"

type KeyValuePair struct {
	Key   string
	Value any
}

type KeyValuePairs []KeyValuePair

func NewKeyValuePair(key string, value any) KeyValuePair {
	return KeyValuePair{Key: key, Value: value}
}

func NewKeyValuePairsFromMap(m map[string]any) KeyValuePairs {
	var pairs KeyValuePairs

	for key, value := range m {
		pairs = append(pairs, NewKeyValuePair(key, value))
	}

	return pairs
}

func (h KeyValuePairs) Add(key string, value any) KeyValuePairs {
	return append(h, NewKeyValuePair(key, value))
}

func (h KeyValuePairs) Sort() KeyValuePairs {
	sort.Slice(h, func(i, j int) bool { return h[i].Key < h[j].Key })

	return h
}
