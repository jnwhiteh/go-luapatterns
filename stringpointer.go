package luapatterns

type sptr struct {
	str []byte
	index int
}

func (s *sptr) clone() *sptr {
	return &sptr{s.str, s.index}
}

func (s *sptr) cloneAt(index int) *sptr {
	return &sptr{s.str, s.index + index}
}

func (s *sptr) getChar() byte {
	return s.getCharAt(0)
}

func (s *sptr) getCharAt(index int) byte {
	i := s.index + index
	if i >= 0 && i < len(s.str) {
		return s.str[i]
	}

	return 0
}

func (s *sptr) postInc(num int) int {
	oldIndex := s.index
	s.index = s.index + num
	return oldIndex
}

func (s *sptr) preInc(num int) int {
	s.index = s.index + num
	return s.index
}

func (s *sptr) length() int {
	return len(s.str) - s.index
}

func (s *sptr) getString() []byte {
	return s.str[s.index:]
}

func (s *sptr) getStringAt(index int) []byte {
	return s.str[s.index + index:]
}

func (s *sptr) getStringLen(length int) []byte {
	end := s.index + length

	if end <= 0 {
		end = len(s.str)
	}

	if end >= len(s.str) {
		return s.str[s.index:]
	} else {
		return s.str[s.index:end]
	}
	panic("never reached")
}
