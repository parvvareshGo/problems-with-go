package main

import "fmt"

func isSorted(numbers []int) bool {
	for i := 0; i < len(numbers)-1; i++ {
		if numbers[i] > numbers[i+1] {
			return false
		}
	}
	return true
}

func main() {
	arr1 := []int{1, 2, 3, 4, 5}
	arr2 := []int{5, 3, 4, 1, 2}

	fmt.Println("Array 1:", arr1, "Is Sorted?", isSorted(arr1))
	fmt.Println("Array 2:", arr2, "Is Sorted?", isSorted(arr2))

}
