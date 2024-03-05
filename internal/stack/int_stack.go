package stack

type intStack struct {
	values []int
}

func NewIntStack() *intStack {
	return &intStack{make([]int, 0)}
}

func (s *intStack) Len() int {
	return len(s.values)
}

func (s *intStack) Peek() int {

	slen := s.Len()
	if slen == 0 {
		return -1
	}

	return s.values[slen-1]
}

func (s *intStack) Pop() int {

	slen := s.Len()
	if slen == 0 {
		return -1
	}

	top := s.values[slen-1]
	s.values = s.values[0 : slen-1]
	return top
}

func (s *intStack) Push(v int) {
	s.values = append(s.values, v)
}
