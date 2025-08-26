package main

import (
	"fmt"
	"strconv"
)

func runLengthEncoding(s string) string {
	if len(s) == 0 {
		return ""
	}

	result := ""
	count := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			count++
		} else {
			result += string(s[i-1]) + strconv.Itoa(count)
		}
	}
	result += string(s[len(s)-1]) + strconv.Itoa(count)
	return result
}

func main() {
	str := "aaabbc"
	encoded := runLengthEncoding(str)
	fmt.Println("Original:", str)
	fmt.Println("Encoded :", encoded)
}
