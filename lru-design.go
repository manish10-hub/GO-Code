package main

import "fmt"

// Doubly Linked List node
type Node struct {
	key, value int
	prev, next *Node
}

// LRU Cache structure
type LRUCache struct {
	capacity int
	cache    map[int]*Node
	head     *Node // MRU
	tail     *Node // LRU
}

// Constructor
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[int]*Node),
	}
}

// Add node to front (MRU)
func (lru *LRUCache) addToHead(node *Node) {
	node.prev = nil
	node.next = lru.head

	if lru.head != nil {
		lru.head.prev = node
	}

	lru.head = node

	if lru.tail == nil { // First node
		lru.tail = node
	}
}

// Remove a node from the list
func (lru *LRUCache) removeNode(node *Node) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		lru.head = node.next // removing head
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		lru.tail = node.prev // removing tail
	}
}

// Move a node to head (MRU)
func (lru *LRUCache) moveToHead(node *Node) {
	lru.removeNode(node)
	lru.addToHead(node)
}

// Remove tail node (LRU)
func (lru *LRUCache) removeTail() *Node {
	oldTail := lru.tail
	lru.removeNode(oldTail)
	return oldTail
}

// Get value for key
func (lru *LRUCache) Get(key int) (int, bool) {
	if node, ok := lru.cache[key]; ok {
		lru.moveToHead(node)
		return node.value, true
	}
	return -1, false
}

// Put key-value pair
func (lru *LRUCache) Put(key, value int) {
	// Case 1: key exists → update + move to head
	if node, ok := lru.cache[key]; ok {
		node.value = value
		lru.moveToHead(node)
		return
	}

	// Case 2: new key
	newNode := &Node{key: key, value: value}
	lru.cache[key] = newNode
	lru.addToHead(newNode)

	// If over capacity → evict LRU
	if len(lru.cache) > lru.capacity {
		tail := lru.removeTail()
		delete(lru.cache, tail.key)
	}
}

func (lru *LRUCache) Print() {
	fmt.Print("LRU state: ")
	curr := lru.head
	for curr != nil {
		fmt.Printf("[%d:%d] ", curr.key, curr.value)
		curr = curr.next
	}
	fmt.Println()
}

func main() {
	lru := NewLRUCache(3)

	lru.Put(1, 10)
	lru.Put(2, 20)
	lru.Put(3, 30)
	lru.Print() // 3 2 1

	_, _ = lru.Get(1)
	lru.Print() // 1 3 2

	lru.Put(4, 40) // evicts 2
	lru.Print()    // 4 1 3

	lru.Put(3, 300) // update 3
	lru.Print()     // 3 4 1
}
