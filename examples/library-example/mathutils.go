package mathutils

import (
	"errors"
	"math"
)

// SafeDivide performs division and returns an error for division by zero
// Demonstrates error propagation with ? operator
func SafeDivide(a float64, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// SafeSqrt calculates square root and returns an error for negative numbers
// Demonstrates parameter type annotations
func SafeSqrt(x float64) (float64, error) {
	if x < 0 {
		return 0, errors.New("cannot calculate square root of negative number")
	}
	return math.Sqrt(x), nil
}

// SafeModulo performs modulo operation and returns an error for zero divisor
// Demonstrates integer type annotations
func SafeModulo(a int, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("modulo by zero")
	}
	return a % b, nil
}

// ChainOperations demonstrates error propagation with ? operator
// Performs: sqrt(a) / b
func ChainOperations(a float64, b float64) (float64, error) {

	__tmp0, __err0 := SafeSqrt(a)

	if __err0 != nil {
		return 0.0, __err0
	}
	// dingo:e:1
	var sqrtA = __tmp0
	__tmp1, __err1 := SafeDivide(sqrtA, b)

	if __err1 != nil {
		return 0.0, __err1
	}
	// dingo:e:1
	var result = __tmp1
	return result, nil
}

// ComputeHypotenuse calculates the hypotenuse of a right triangle
// Demonstrates chained ? operators
func ComputeHypotenuse(sideA float64, sideB float64) (float64, error) {
	__tmp0, __err0 := SafeMultiply(sideA, sideA)

	if __err0 != nil {
		return 0.0, __err0
	}
	// dingo:e:1
	var aSq = __tmp0
	__tmp1, __err1 := SafeMultiply(sideB, sideB)

	if __err1 != nil {
		return 0.0, __err1
	}
	// dingo:e:1
	var bSq = __tmp1
	sum := aSq + bSq
	__tmp2, __err2 := SafeSqrt(sum)

	if __err2 != nil {
		return 0.0, __err2
	}
	// dingo:e:1
	var result = __tmp2
	return result, nil
}

// SafeMultiply is a helper that never fails but matches the pattern
func SafeMultiply(a float64, b float64) (float64, error) {
	return a * b, nil
}

// Percentage calculates what percentage a is of b
// Returns error if b is zero
func Percentage(a float64, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("cannot calculate percentage of zero")
	}
	return (a / b) * 100, nil
}

// Average calculates the average of a slice of numbers
// Returns error if slice is empty
func Average(numbers []float64) (float64, error) {
	if len(numbers) == 0 {
		return 0, errors.New("cannot calculate average of empty slice")
	}

	sum := 0.0
	for _, num := range numbers {
		sum += num
	}

	return sum / float64(len(numbers)), nil
}
