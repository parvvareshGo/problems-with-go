package main

import "fmt"

func CountingElementFrequencies(number []int) map[int]int {
	counter := make(map[int]int)
	for _, element := range number {
		counter[element]++
	}
	return counter
}

func main() {
	arr := []int{1, 2, 2, 3, 4, 3, 2, 5, 1, 4, 4}
	result := CountingElementFrequencies(arr)
	for key, value := range result {

		fmt.Printf("%d -> %d\n", key, value)
	}
}
