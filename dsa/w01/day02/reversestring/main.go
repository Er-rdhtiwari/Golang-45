package main

import (
	"fmt"
	"strings"

)


func ReverseStringBruteForce(s string) string{
	result := ""

	for i:= len(s)-1;i>=0;i--{
		result += string(s[i])
	}
	return result
}


func ReverseString(s string) string{
	var builder strings.Builder

	for i:= len(s)-1; i>=0; i--{
		builder.WriteByte(s[i])
	}
	return builder.String()
}

func ReverseStringUnicode(s string) string{
	runes := []rune(s)
	left :=0
	right := len(runes)-1
	for left< right {
		runes[left], runes[right] = runes[right], runes[left]
		left++
		right--
	}
	return string(runes)
}

func main() {
    fmt.Println(ReverseStringBruteForce("hello"))
	fmt.Println(ReverseString("Eliphant"))
	fmt.Println(ReverseStringUnicode("hello"))
    fmt.Println(ReverseStringUnicode("😊a"))
}