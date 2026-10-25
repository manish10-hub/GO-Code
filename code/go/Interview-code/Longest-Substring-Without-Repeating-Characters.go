// Online Go compiler to run Golang program online
// Print "Start small. Ship something." message

package main

import "fmt"

func main() {
	input := "abcabcdebb"
	fmt.Println(longestString(input))
}

func longestString(input string) int {
	uniqueMap := make(map[byte]bool, len(input))
	maxLen := 0
	left := 0

	for right := 0; right < len(input); right++ {
		for uniqueMap[input[right]] {
			delete(uniqueMap, input[left])
			left++
		}
		uniqueMap[input[right]] = true
		currLen := right - left + 1
		if maxLen < currLen {
			maxLen = currLen
		}

	}
	return maxLen

}
