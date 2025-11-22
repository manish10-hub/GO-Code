package main

import "fmt"

func main() {
	graph := map[int][]int{
		0: {1, 2},
		1: {0, 3},
		2: {0},
		3: {1},
	}

	fmt.Println("DFS traversal:")
	//visited := make(map[int]bool)
	dfs(0, graph)

	fmt.Println("BFS traversal:")
	bfs(0, graph)
}

func bfs(node int, graph map[int][]int) {

	queue := []int{node}
	visitedNode := make(map[int]bool)
	visitedNode[node] = true

	for len(queue) > 0 {
		data := queue[0]
		queue = queue[1:]
		fmt.Println("visited ", data)

		for _, neighbour := range graph[data] {
			if !visitedNode[neighbour] {
				visitedNode[neighbour] = true
				queue = append(queue, neighbour)
			}
		}
	}
}

func dfs(node int, graph map[int][]int) {

	visitedNode := make(map[int]bool)
	var helper func(node int, graph map[int][]int, visitedNode map[int]bool)
	helper = func(node int, graph map[int][]int, visitedNode map[int]bool) {
		if visitedNode[node] {
			return
		}
		visitedNode[node] = true
		fmt.Println("visited ", node)
		for _, neighbour := range graph[node] {
			helper(neighbour, graph, visitedNode)
		}
	}
	helper(node, graph, visitedNode)
}

func hasCycleDFS(node int, parent int, graph map[int][]int, visited map[int]bool) bool {
	// Mark the current node as visited
	visited[node] = true

	// Traverse all neighbors of the current node
	for _, neighbour := range graph[node] {
		if !visited[neighbour] {
			// Recurse on unvisited neighbor
			if hasCycleDFS(neighbour, node, graph, visited) {
				return true // Cycle found in deeper call
			}
		} else if neighbour != parent {
			// If neighbor is visited and not parent → cycle exists
			return true
		}
	}

	// No cycle detected from this node
	return false
}

// For directed Graph
func hasCycle(node int, graph map[int][]int) bool {
	visited := make(map[int]bool)
	track := make(map[int]bool)

	var helper func(node int, graph map[int][]int, visited, track map[int]bool) bool
	helper = func(node int, graph map[int][]int, visited, track map[int]bool) bool {

		visited[node] = true
		track[node] = true

		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				if helper(neighbor, graph, visited, track) {
					return true
				}
			} else {
				if track[neighbor] {
					return true
				}
			}
		}
		track[node] = false
		return false
	}
	return helper(node, graph, visited, track)
}

func isolatedGraph(graph map[int][]int) int {
	visited := make(map[int]bool)
	count := 0

	var helper func(node int, graph map[int][]int, visited map[int]bool)
	helper = func(node int, graph map[int][]int, visited map[int]bool) {

		for _, neighbour := range graph[node] {
			if !visited[neighbour] {
				visited[neighbour] = true
				helper(neighbour, graph, visited)
			}
		}

	}

	for node := range graph {
		if !visited[node] {
			visited[node] = true
			count++
			helper(node, graph, visited)
		}

	}
	return count
}
