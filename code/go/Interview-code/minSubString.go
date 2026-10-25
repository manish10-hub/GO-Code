// Online Go compiler to run Golang program online
// Print "Start small. Ship something." message

package main

import "fmt"

func main() {
	input := "abcabcdebb"
	target := "ad"
	fmt.Println(minSubString(input, target))
}

func isExist(input, target map[byte]int) bool {
	for key, value := range target {
		if input[key] < value {
			return false
		}
	}
	return true
}

func minSubString(input, target string) string {
	targetFreqMap := make(map[byte]int)

	for index := 0; index < len(target); index++ {
		targetFreqMap[target[index]]++
	}

	inputFreqMap := make(map[byte]int)

	maxLen := len(input) + 1
	left := 0
	minString := ""

	for right := 0; right < len(input); right++ {
		inputFreqMap[input[right]]++

		for isExist(inputFreqMap, targetFreqMap) {
			if maxLen > (right - left + 1) {
				maxLen = right - left + 1
				minString = input[left : right+1]
			}
			inputFreqMap[input[left]]--
			left++
		}

	}
	return minString

}
