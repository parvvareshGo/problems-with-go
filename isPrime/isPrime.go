package main

import (
	"fmt"
)

func isPrime(number int) bool {
	if number < 2 {
		return false
	}
	for i := 2; i*i <= number; i++ {
		if number%i == 0 {
			return false
		}
	}
	return true
}

func main() {
	n := 5
	fmt.Printf("Is %d prime? %v\n", n, isPrime(n))
}
