package main

import (
	"fmt"
)

func removeDuplicates(numbers []int) []int {
	seen := make(map[int]bool)
	result := []int{}

	for _, num := range numbers {
		if !seen[num] {
			seen[num] = true
			result = append(result, num)

		}

	}
	return result
}

func main() {
	arr := []int{1, 2, 2, 3, 4, 3, 5, 1}
	fmt.Println("Original array:", arr)
	fmt.Println("Array without duplicates:", removeDuplicates(arr))
}
