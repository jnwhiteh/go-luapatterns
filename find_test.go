package luapatterns

import (
	"testing"
)

type Test struct {
	s1 string
	s2 string
	succ bool
	start int
	end int
}

var PlainTests = []Test{
	Test{"", "", true, 0, 0},
	Test{"a", "a", true, 0, 1},
	Test{"a", "b", false, -1, -1},
	Test{"ab", "b", true, 1, 2},
	Test{"ab", "a", true, 0, 1},
	Test{"aaa", "aaa", true, 0, 3},
	Test{"aaabaa", "aaa", true, 0, 3},
	Test{"aaabaa", "baa", true, 3, 6},
	Test{"aaa", "b", false, -1, -1},
	Test{"aaaba", "baa", false, -1, -1},
	Test{"aaabbaba", "aba", true, 5, 8},
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
