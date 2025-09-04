package main

import (
	"fmt"
)

func bubbleSort(arr []int){
	n := len(arr)

	if n < 2 {
		return
	}
	for pass := 0; pass < n-1; pass++ {
		swapped := false


		for i := 0 ; i < n-1-pass; i++{
			if arr[i] > arr[i + 1]{
				arr[i], arr[i+1] = arr[i+1], arr[i]
				swapped = true	
			}
		}
		if !swapped {
			break
		}
	}
}

func main() {
	a := []int{5, 1, 4, 2, 8, 0, 2}
	fmt.Println("Before:", a)
	bubbleSort(a)
	fmt.Println("After :", a)
}
