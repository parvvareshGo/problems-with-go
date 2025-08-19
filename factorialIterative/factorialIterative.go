package main

import (
	"fmt"
)

func factorialIterative(n int) int {
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result
}

func main() {
	n := 5
	fmt.Printf("Factorial of %d (iterative) = %d\n", n, factorialIterative(n))
}
