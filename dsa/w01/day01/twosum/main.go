package main

import "fmt"

func TwoSumBruteForce(nums []int,target int) []int{
	for i:=0; i<len(nums); i++{
		for j:= i+1; j < len(nums); j++{
			if nums[i]+nums[j] == target{
				return []int{i,j}
			}
		}
	}
	return []int{}
}

func TwoSum(nums []int, target int) []int{
	seen := make(map[int]int)
	for i, value := range nums {
		needed := target- value

		if prviousIndex, ok := seen[needed]; ok {
				return []int{prviousIndex, i}
			}
		seen[value] = i
	}
	return []int{}
}

func main(){
	nums := []int{2,7,11,15}
	target := 9
	// result := TwoSumBruteForce(nums, target)
	result := TwoSum(nums, target)
	fmt.Println(result)
}