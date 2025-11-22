package main

type IntStack struct {
	data []int
}

// Push adds an element to the top of the stack.
func (s *IntStack) Push(value int) {
	s.data = append(s.data, value)
}

// Pop removes and returns the top element of the stack.
func (s *IntStack) Pop() (int, bool) {
	if len(s.data) == 0 {
		return 0, false
	}
	val := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return val, true
}

// Peek returns the top element without removing it.
func (s *IntStack) Peek() (int, bool) {
	if len(s.data) == 0 {
		return 0, false
	}
	return s.data[len(s.data)-1], true
}

// IsEmpty checks if the stack is empty.
func (s *IntStack) IsEmpty() bool {
	return len(s.data) == 0
}
