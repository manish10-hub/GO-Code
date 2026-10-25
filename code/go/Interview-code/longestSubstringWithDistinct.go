// Online Go compiler to run Golang program online
// Print "Start small. Ship something." message

package main

import "fmt"

func main() {
	input := "abcabcdebb"
	target := 3
	fmt.Println(longestSubstring(input, target))
}

func longestSubstring(input string, distinct int) string {
	freqMap := make(map[byte]int)
	left := 0
	result := ""

	for right := 0; right < len(input); right++ {
		freqMap[input[right]]++
		for len(freqMap) > distinct {
			freqMap[input[left]]--
			if freqMap[input[left]] == 0 {
				delete(freqMap, input[left])
			}
			left++
		}
		if len(result) < right-left+1 {
			result = input[left : right+1]
		}
	}
	return result

}
