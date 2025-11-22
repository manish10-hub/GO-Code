// Online Go compiler to run Golang program online
// Print "Try programiz.pro" message

package main

import (
	"fmt"
	"math"
)

type TreeNode struct {
	Value int
	Left  *TreeNode
	Right *TreeNode
}

func InOrderTraversal(root *TreeNode) {
	if root == nil {
		return
	}
	InOrderTraversal(root.Left)
	fmt.Println(root.Value)
	InOrderTraversal(root.Right)
}

// LevelOrderTraversal performs level-order traversal (Breadth-First Search).
func LevelOrderTraversal(root *TreeNode) {
	if root == nil {
		return
	}
	queue := []*TreeNode{root}

	for len(queue) > 0 {
		treeNode := queue[0]
		fmt.Println(treeNode.Value)
		if treeNode.Left != nil {
			queue = append(queue, treeNode.Left)
		}
		if treeNode.Right != nil {
			queue = append(queue, treeNode.Right)
		}
		queue = queue[1:]
	}
}

func hasAPath(root *TreeNode, sum int) bool {
	if root == nil {
		return false
	}
	if root.Left == nil && root.Right == nil {
		return root.Value == sum
	}
	return (hasAPath(root.Left, sum-root.Value) || hasAPath(root.Right, sum-root.Value))

}

func lcm(root, p, q *TreeNode) *TreeNode {
	if root == nil || p == root || q == root {
		return root
	}
	left := lcm(root.Left, p, q)
	right := lcm(root.Right, p, q)
	if left != nil && right != nil {
		return root
	}
	if left != nil {
		return left
	}
	return right

}

func maxDepth(root *TreeNode) int {
	if root == nil {
		return 0
	}
	return max(maxDepth(root.Left), maxDepth(root.Right)) + 1
}

// Checks if tree is a valid Binary Search Tree
func bstCheck(root *TreeNode) bool {
	return bstCheck2(root, math.MinInt64, math.MaxInt64)
}

// Recursively ensures that all node values are within valid range
func bstCheck2(root *TreeNode, min, max int64) bool {
	if root == nil {
		return true
	}
	val := int64(root.Value)
	if val <= min || val >= max {
		return false
	}
	return bstCheck2(root.Left, min, val) &&
		bstCheck2(root.Right, val, max)
}

func maxSum(root *TreeNode) int {
	if root == nil {
		return 0
	}
	maxSum := math.MinInt64
	var helper func(*TreeNode) int
	helper = func(root *TreeNode) int {
		if root == nil {
			return 0
		}
		if root.Left == nil && root.Right == nil {
			return root.Value
		}
		leftMax := helper(root.Left)
		rightMax := helper(root.Right)
		maxSum = max(maxSum, (leftMax + rightMax + root.Value))
		return max(leftMax, rightMax) + root.Value
	}
	helper(root)
	return maxSum
}

func main() {
	// Constructing the binary tree:
	//         1
	//        / \
	//       2   3
	//      / \   \
	//     4   5   6

	root := &TreeNode{Value: 1}
	root.Left = &TreeNode{Value: 2}
	root.Right = &TreeNode{Value: 3}
	root.Left.Left = &TreeNode{Value: 4}
	root.Left.Right = &TreeNode{Value: 5}
	root.Right.Right = &TreeNode{Value: 6}

	fmt.Print("In-order Traversal: ")
	InOrderTraversal(root)
	fmt.Println()

	/*fmt.Print("Pre-order Traversal: ")
	PreOrderTraversal(root)
	fmt.Println()

	fmt.Print("Post-order Traversal: ")
	PostOrderTraversal(root)
	fmt.Println()*/

	fmt.Print("Level-order Traversal: ")
	LevelOrderTraversal(root)
	fmt.Println()

}
