package utils

import (
	"fmt"
	"strconv"

	"golang.org/x/exp/constraints"
)

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func ZeroPadInt64(number int64) string {
	return fmt.Sprintf("%0*d", strconv.IntSize/4, number)
}
