package core

func IntRef(value int) *int {
	return &value
}

func Int64Ref(value int64) *int64 {
	return &value
}

func Map[T, V any](arr []T, f func(T) V) []V {
	result := make([]V, len(arr))
	for i := range arr {
		result[i] = f(arr[i])
	}
	return result
}

func Filter[T any](arr []T, f func(T) bool) []T {
	var result []T

	for i := range arr {
		if f(arr[i]) {
			result = append(result, arr[i])
		}
	}
	return result
}

func Contains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
