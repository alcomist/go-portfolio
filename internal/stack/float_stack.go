package stack

type floatStack struct {
	values []float64
}

func NewFloatStack() *floatStack {
	return &floatStack{make([]float64, 0)}
}

func (s *floatStack) Len() int {
	return len(s.values)
}

func (s *floatStack) Peek() float64 {

	slen := s.Len()
	if slen == 0 {
		return -1
	}

	return s.values[slen-1]
}

func (s *floatStack) Pop() float64 {

	slen := s.Len()
	if slen == 0 {
		return -1
	}

	top := s.values[slen-1]
	s.values = s.values[0 : slen-1]
	return top
}

func (s *floatStack) Push(v float64) {
	s.values = append(s.values, v)
}
