// find the top k frequent elements in an array

//Pattern: Bucket Sort + frequency map

package main

import "fmt"

func topKElement(arr []int, k int) []int {
	freqMap := make(map[int]int)

	for _, num := range arr {
		freqMap[num]++
	}

	bucket := make([][]int, len(arr)+1)

	for num, freq := range freqMap {
		bucket[freq] = append(bucket[freq], num)
	}
	var result []int

	for last := len(bucket) - 1; last >= 0; last-- {
		myarr := bucket[last]
		for i := 0; i < len(myarr); i++ {
			result = append(result, myarr[i])
		}

	}
	return result[:k]
}

func main() {
	arr := []int{1, 3, 4, 2, 3, 4, 5, 5, 5, 2, 4, 1, 1, 1}
	fmt.Println(topKElement(arr, 3))
}
