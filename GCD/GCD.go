package main 

import "fmt"

func gcd(a, b int) int {
	if b == 0 {
		return a
	}
	return gcd(b , a % b)
}

func main() {
	fmt.Println("GCD of 48 and 18 is:", gcd(48, 18))
}
