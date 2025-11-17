package main

import "fmt"

func main() {
	numbers := []int{1, 2, 3, 4, 5}

	// This will be transformed by the plugin
	doubled := numbers.map(func(x int) int { return x * 2 })
	fmt.Println("Doubled:", doubled)

	// Filter
	evens := numbers.filter(func(x int) bool { return x%2 == 0 })
	fmt.Println("Evens:", evens)

	// Reduce
	sum := numbers.reduce(0, func(acc int, x int) int { return acc + x })
	fmt.Println("Sum:", sum)

	// Sum helper
	total := numbers.sum()
	fmt.Println("Total:", total)

	// Count
	count := numbers.count(func(x int) bool { return x > 2 })
	fmt.Println("Count > 2:", count)

	// All
	all := numbers.all(func(x int) bool { return x > 0 })
	fmt.Println("All positive:", all)

	// Any
	any := numbers.any(func(x int) bool { return x < 0 })
	fmt.Println("Has negative:", any)

	// Chaining
	result := numbers.filter(func(x int) bool { return x > 2 }).map(func(x int) int { return x * 10 })
	fmt.Println("Chained:", result)
}
