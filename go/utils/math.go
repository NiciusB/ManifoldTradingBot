package utils

import "golang.org/x/exp/constraints"

type Number interface {
	constraints.Integer | constraints.Float
}

// Clamp returns num clamped to [low, high]
func Clamp[T Number](num T, low T, high T) T {
	if num < low {
		return low
	}
	if num > high {
		return high
	}
	return num
}

// Map from [sourceLow, sourceHigh] to [targetLow, targetHigh] linearly
func MapNumber[T Number](num T, sourceLow T, sourceHigh T, targetLow T, targetHigh T) T {
	var mapped = ((num-sourceLow)*(targetHigh-targetLow))/(sourceHigh-sourceLow) + targetLow
	return Clamp(mapped, targetLow, targetHigh)
}
