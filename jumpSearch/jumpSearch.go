package main

import (
	"fmt"
	"math"
)

func jumpSearch(arr []int, target int) int{
	n := len(arr)
	step := int(math.Sqrt(float64(n))) 
	prev := 0

	for arr[int(math.Min(float64(step), float64(n)-1))] < target{
		prev = step
		step += int(math.Sqrt(float64(n)))
		if prev >= n {
			return -1
		}	
	}

	for arr[prev] < target{
		prev++
		if prev == int(math.Min(float64(step), float64(n))) {
			return -1
		}
	}

	if arr[prev] == target{
		return prev
	}

	return -1
}

func main() {
	arr := []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21}
	target := 15

	index := jumpSearch(arr, target)
	if index != -1 {
		fmt.Printf("Element %d found at index %d\n", target, index)
	} else {
		fmt.Printf("Element %d not found\n", target)
	}
}
