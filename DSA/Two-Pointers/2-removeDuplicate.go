// remove duplicates from a sorted array

//pattern: Two Pointers

package main

import "fmt"

func removeDuplicates(arr []int) []int {
	if len(arr) == 0 {
		return arr
	}

	left := 0

	for right := 1; right < len(arr); right++ {
		if arr[right] != arr[left] {
			left++
			arr[left] = arr[right]
		}
		fmt.Println(arr)
	}

	return arr[:left+1]
}

func main() {
	arr := []int{1, 1, 3, 3, 4, 6, 6, 8, 9, 9}
	fmt.Println(removeDuplicates(arr))
}
