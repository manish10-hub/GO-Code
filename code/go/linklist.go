// Online Go compiler to run Golang program online
// Print "Try programiz.pro" message

package main

import "fmt"

type Node struct {
	data int
	next *Node
}

type LinkedList struct {
	head *Node
}

func (ll *LinkedList) InsertAtEnd(val int) {
	fmt.Println("data is ", val)
	node := &Node{data: val, next: nil}
	if ll.head == nil {
		ll.head = node
		return
	}
	current := ll.head
	for current.next != nil {
		current = current.next
	}
	current.next = node
}
func (ll *LinkedList) DeleteAtFront() {
	if ll.head == nil {
		return
	}
	current := ll.head
	ll.head = current.next
}
func (ll *LinkedList) DeleteAtEnd() {
	if ll.head == nil {
		return
	}
	current := ll.head
	for current.next.next != nil {
		current = current.next
	}
	current.next = nil
}
func (ll *LinkedList) printLL() {
	current := ll.head
	for current != nil {
		fmt.Println(current.data)
		current = current.next
	}
}

func (ll *LinkedList) ReverseLL() {
	if ll.head == nil || ll.head.next == nil {
		return
	}

	prev := ll.head
	current := prev.next
	prev.next = nil

	for current.next != nil {
		next := current.next
		current.next = prev
		prev = current
		current = next
	}

	current.next = prev
	ll.head = current
}
func mergeSortedLL(head1, head2 *Node) *Node {
	dummy := &Node{}
	tail := dummy

	p1, p2 := head1, head2

	for p1 != nil && p2 != nil {
		// Debug print (can be commented out in production)
		fmt.Printf("p1: %d, p2: %d\n", p1.data, p2.data)

		if p1.data < p2.data {
			tail.next = p1
			p1 = p1.next
		} else {
			tail.next = p2
			p2 = p2.next
		}
		tail = tail.next
	}

	// Attach remaining part
	if p1 != nil {
		tail.next = p1
	} else {
		tail.next = p2
	}

	return dummy.next
}
func (ll *LinkedList) removeDup() {
	if ll.head == nil {
		return
	}

	prev := ll.head
	myMap := make(map[int]bool)
	myMap[prev.data] = true

	current := prev.next
	for current != nil {
		if _, ok := myMap[current.data]; ok {
			// Duplicate: remove node
			prev.next = current.next
		} else {
			// Unique: track and move prev
			myMap[current.data] = true
			prev = current
		}
		current = prev.next
	}
}

func main() {
	ll := &LinkedList{}
	ll.InsertAtEnd(10)
	ll.InsertAtEnd(5)
	ll.InsertAtEnd(35)
	ll.InsertAtEnd(20)
	ll.InsertAtEnd(25)
	ll.printLL()
	/*ll.DeleteAtFront()
	  ll.DeleteAtFront()*/
	ll.DeleteAtEnd()
	fmt.Println("manish")
	ll.printLL()
	fmt.Println("Try programiz.pro")
}
