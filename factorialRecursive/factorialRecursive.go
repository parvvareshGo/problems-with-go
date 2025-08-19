package main

import (
	"fmt"
)

func factorialRecursive(number int) int {
	if number == 0 || number == 1 {
		return 1
	}
	return number * factorialRecursive(number-1)

}

func main() {
	n := 5
	fmt.Printf("Factorial of %d (recursive) = %d\n", n, factorialRecursive(n))
}
