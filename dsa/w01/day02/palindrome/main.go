package main

import "fmt"

func IsPalindrome(s string) bool{
	left :=0
	right := len(s)-1

	for left<right{
		if s[left] != s[right]{
			return false
		}
		left++
		right--

	}
	return true
}

func main() {
    fmt.Println(IsPalindrome("madam"))
    fmt.Println(IsPalindrome("racecar"))
    fmt.Println(IsPalindrome("hello"))
}