package main

import "fmt"

func MaxMin(numbers []int) {
	max := numbers[0]
	min := numbers[0]
	for _, value := range numbers {
		if value > max {
			max = value
		}
		if value < min {
			min = value
		}
	}

	fmt.Println("Maximum:", max)
	fmt.Println("Minimum:", min)
}
func main() {
	arr := []int{5, 2, 9, 1, 7, 6}
	MaxMin(arr)

}
