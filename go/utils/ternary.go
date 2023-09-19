package utils

func Ternary[T any](statement bool, value1 T, value2 T) T {
	if statement {
		return value1
	} else {
		return value2
	}
}
