// Online Go compiler to run Golang program online
// Print "Start small. Ship something." message

package main

import "fmt"

func twoSum(arr []int, sum int) bool {
	l := len(arr)
	left := 0
	right := l - 1
	for left < right {
		currSum := arr[left] + arr[right]
		if currSum == sum {
			return true
		}
		if currSum < sum {
			left++
		} else {
			right--
		}
	}
	return false
}

func main() {
	arr := []int{1, 3, 4, 6, 8, 9}
	fmt.Println(twoSum(arr, 18))
}
