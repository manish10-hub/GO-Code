// Online Go compiler to run Golang program online
// Print "Start small. Ship something." message

package main

import (
	"fmt"
	"sort"
)

type Interval struct {
	Start int
	End   int
}

func mergeIntervals(intervals []Interval) []Interval {

	if len(intervals) <= 1 {
		return intervals
	}
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i].Start < intervals[j].Start
	})
	result := []Interval{intervals[0]}

	for i := 1; i < len(intervals); i++ {
		current := intervals[i]
		last := &result[len(result)-1]
		if current.Start <= last.End {
			last.End = current.End
		} else {
			result = append(result, current)
		}
	}
	return result
}

func main() {

	intervals := []Interval{
		{1, 3},
		{2, 6},
		{8, 10},
		{15, 18},
	}

	fmt.Println(mergeIntervals(intervals))
}
