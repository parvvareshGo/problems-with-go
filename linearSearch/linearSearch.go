package main

import (
	"fmt"
)

func linearSearch(arr []int, target int) int{
	for i, v := range(arr){
		if v == target{
			return i
		}
	}
	return -1
}

func main(){
	arr := []int{10, 25, 30, 45, 50}
	target := 30

	index := linearSearch(arr, target)
	if index != -1 {
		fmt.Printf("Element %d found at index %d\n", target, index)
	} else {
		fmt.Printf("Element %d not found\n", target)
	}
}
