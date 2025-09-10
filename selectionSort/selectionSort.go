package main

import (
	"fmt"
)


func selectionSort(arr []int) {
	n := len(arr)
	for i := 0; i < n - 1; i++ {
		minIndex := i
		for j := i; j < n ; j++ {
			if arr[j] < arr[minIndex] {
				minIndex = j
			}
		}

		arr[i], arr[minIndex] = arr[minIndex], arr[i]
	}
}


func main() {
	arr := []int{64, 25, 12, 22, 11}
	fmt.Println("Before sorting:", arr)

	selectionSort(arr)

	fmt.Println("After sorting:", arr)
}
