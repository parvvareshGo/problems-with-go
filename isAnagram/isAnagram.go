package main

import "fmt"

func isAnagram(s1, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}
	count := make(map[rune]int)

	for _, ch := range s1 {
		count[ch]++
	}

	for _, ch := range s2 {
		count[ch]--
		if count[ch] < 0 {
			return false
		}

	}
	return true
}

func main() {
	fmt.Println(isAnagram("listen", "silent")) // true
	fmt.Println(isAnagram("hello", "world"))   // false
}
