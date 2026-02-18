package utils

func Map[T, U any](source []T, transform func(T) U) []U {
	if source == nil {
		return nil
	}

	result := make([]U, len(source))

	for i, v := range source {
		result[i] = transform(v)
	}

	return result
}

func MapWithIndex[T, U any](source []T, transform func(int, T) U) []U {
	if source == nil {
		return nil
	}

	result := make([]U, len(source))

	for i, v := range source {
		result[i] = transform(i, v)
	}

	return result
}

func Filter[T any](source []T, predicate func(T) bool) []T {
	if source == nil {
		return nil
	}

	result := make([]T, 0, len(source))

	for _, v := range source {
		if predicate(v) {
			result = append(result, v)
		}
	}

	return result
}

func FilterMap[T, U any](source []T, predicate func(T) bool, transform func(T) U) []U {
	if source == nil {
		return nil
	}

	result := make([]U, 0, len(source))

	for _, v := range source {
		if predicate(v) {
			result = append(result, transform(v))
		}
	}

	return result
}
