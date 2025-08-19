package main

import (
	"fmt"
	"strings"
)

func isPalindrome(text string) bool {
	text = strings.ToLower(text)
	left, right := 0, len(text)-1

	for left < right {
		if text[left] != text[right] {
			return false
		}
		left++
		right--

	}
	return true
}

func main() {
	word := "assa"
	fmt.Printf("%s -> %v\n", word, isPalindrome(word))

}
