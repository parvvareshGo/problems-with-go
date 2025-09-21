package main

import (
	"fmt"
)

func hash(key string, size int) int{
	sum := 0
	for _, ch := range key {
		sum += int(ch)
	}
	return sum % size
}

func main() {
	size := 10

	names := []string{"Ali", "Sara", "Mona", "Omid", "Nima"}
	for _ , name := range names {
		index := hash(name, size)
		fmt.Printf("%s -> index is  %d\n", name, index)

	}
}
