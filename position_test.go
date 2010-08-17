package luapatterns

import (
	"testing"
)

type posTest struct {
	str string
	pat string
	init int
	succ bool
	start int
	end int
}

// These tests have been taken from the Lua test suite, but have been altered
// to reflect the difference in array indexing and substrings (i.e) str[1:1]
// is always an empty string. As a result the start and end returns must be
// adjusted.
var posTests = []posTest{
	posTest{"", "", 0, true, 0, 0},									// special case
	posTest{"alo", "", 0, true, 0, 0},								// special case
	posTest{"a\x00o a\x00o a\x00o", "a", 0, true, 0, 1},
	posTest{"a\x00o a\x00o a\x00o", "a\x00o", 2, true, 4, 7},
	posTest{"a\x00o a\x00o a\x00o", "a\x00o", 8, true, 8, 11},
	posTest{"a\x00oa\x00a\x00a\x00\x00ab", "\x00ab", 1, true, 9, 12},
	posTest{"a\x00oa\x00a\x00a\x00\x00ab", "b", 0, true, 11, 12},
	posTest{"a\x00oa\x00a\x00a\x00\x00ab", "b\x00", 0, false, 0, 0},
	posTest{"", "\x00", 0, false, 0, 0},
	posTest{"alo123alo", "12", 0, true, 3, 5},
	posTest{"alo123alo", "^12", 0, false, 0, 0},
}

func TestPatternPos(t *testing.T) {
	for _, test := range posTests {
		succ, start, end, _ := Find(test.str[test.init:], test.pat, false)
		start = start + test.init
		end = end + test.init

		if succ != test.succ {
			t.Errorf("find('%s', '%s', %d) returned %t, expected %t", test.str, test.pat, test.init, succ, test.succ)
		}
		if succ && (start != test.start || end != test.end) {
			t.Errorf("find('%s', '%s', %d) returned (%d, %d), expected (%d, %d)", test.str, test.pat, test.init, start, end, test.start, test.end)
		}
	}
}
