package main

import (
	"fmt"
)

func binarySearch(arr []int, target int) int{
	left, right := 0, len(arr) - 1
	for left <= right{
		mid := left + (right - left) / 2
		if arr[mid] == target{
			return mid
		} else if arr[mid] < target{
			left = mid + 1
		} else {
		right = mid - 1
		}

	}
	return -1
}



func main() {
	arr := []int{1, 3, 5, 7, 9, 11}
	target := 7

	index := binarySearch(arr, target)
	if index != -1 {
		fmt.Printf("Element %d found at index %d\n", target, index)
	} else {
		fmt.Printf("Element %d not found\n", target)
	}
}
