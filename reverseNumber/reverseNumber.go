package main

import "fmt"

func reverseNumber(number int) int {
	rev := 0
	for number != 0 {
		temp := number % 10
		rev = (rev * 10) + temp
		number = number / 10

	}
	return rev
}

func main() {
	num := 1234
	fmt.Printf("this is rev of number: %d", reverseNumber(num))
}
