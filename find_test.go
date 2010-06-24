package luapatterns

import (
	"testing"
)

type PlainTest struct {
	s1 string
	s2 string
	succ bool
	start int
	end int
}

var PlainTests = []PlainTest{
	PlainTest{"", "", true, 0, 0},
	PlainTest{"a", "a", true, 0, 1},
	PlainTest{"a", "b", false, -1, -1},
	PlainTest{"ab", "b", true, 1, 2},
	PlainTest{"ab", "a", true, 0, 1},
	PlainTest{"aaa", "aaa", true, 0, 3},
	PlainTest{"aaabaa", "aaa", true, 0, 3},
	PlainTest{"aaabaa", "baa", true, 3, 6},
	PlainTest{"aaa", "b", false, -1, -1},
	PlainTest{"aaaba", "baa", false, -1, -1},
	PlainTest{"aaabbaba", "aba", true, 5, 8},
}

func TestPlainFind(t *testing.T) {
	for _, test := range PlainTests {
		s1 := []byte(test.s1)
		s2 := []byte(test.s2)
		start := lmemfind(s1, s2)
		if start != test.start{
			t.Errorf("Fail in lmemfind('%s', '%s') => %d instead of %d\n",
				s1, s2, start, test.start)
		}

		succ, start, end := str_find_aux(s1, s2, 0, true)
		if succ != test.succ || start != test.start || end != test.end {
			t.Errorf("Fail in str_find_aux('%s', '%s', 0, true) => ('%t', '%d', '%d') instead of ('%t', '%d', '%d')\n",
				s1, s2, succ, start, end, test.succ, test.start, test.end)
		}
	}
}
