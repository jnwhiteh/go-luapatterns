package luapatterns

import (
	"fmt"
	"testing"
)

type subTest struct {
	str  string
	pat  string
	succ bool
	cap  string
}

var subTests = []subTest{
	subTest{"aaab", "a*", true, "aaa"},
	subTest{"aaa", "^.*$", true, "aaa"},
	subTest{"aaa", "b*", true, ""},
	subTest{"aaa", "ab*a", true, "aa"},
	subTest{"aba", "ab*a", true, "aba"},
	subTest{"aaab", "a+", true, "aaa"},
	subTest{"aaa", "^.+$", true, "aaa"},
	subTest{"aaa", "b+", false, ""},
	subTest{"aaa", "ab+a", false, ""},
	subTest{"aba", "ab+a", true, "aba"},
	subTest{"a$a", ".$", true, "a"},
	subTest{"a$a", ".%$", true, "a$"},
	subTest{"a$a", ".$.", true, "a$a"},
	subTest{"a$a", "$$", false, ""},
	subTest{"a$b", "a$", false, ""},
	subTest{"a$a", "$", true, ""},
	subTest{"", "b*", true, ""},
	subTest{"aaa", "bb*", false, ""},
	subTest{"aaab", "a-", true, ""},
	subTest{"aaa", "^.-$", true, "aaa"},
	subTest{"aabaaabaaabaaaba", "b.*b", true, "baaabaaabaaab"},
	subTest{"aabaaabaaabaaaba", "b.-b", true, "baaab"},
	subTest{"alo xo", ".o$", true, "xo"},
	subTest{" \n isto \x82 assim", "%S%S*", true, "isto"},
	subTest{" \n isto \x82 assim", "%S*$", true, "assim"},
	subTest{" \n isto \x82 assim", "[a-z]*$", true, "assim"},
	subTest{"im caracter ? extra", "[^%sa-z]", true, "?"},
	subTest{"", "a?", true, ""},

	// These tests don't work 100% correctly if you use Unicode á instead
	// of the ASCII value \225. In particular the third test fails
	subTest{"\225", "\225?", true, "\225"},
	subTest{"\225bl", "\225?b?l?", true, "\225bl"},
	subTest{"  \225bl", "\225?b?l?", true, ""},

	subTest{"aa", "^aa?a?a", true, "aa"},
	subTest{"]]]\225b", "[^]]", true, "\225"},
	subTest{"0alo alo", "%x*", true, "0a"},
	subTest{"alo alo", "%C+", true, "alo alo"},

	// These are grouped seperately in the original tests
	// TODO: re-enable this test
	subTest{"alo alx 123 b\x00o b\x00o", "(..*) %1", true, "b\x00o b\x00o"},
	subTest{"axz123= 4= 4 34", "(.+)=(.*)=%2 %1", true, "3= 4= 4 3"},
}

func TestSubtring(t *testing.T) {
	enableDebug = false

	for _, test := range subTests {
		debug(fmt.Sprintf("=== Find(%q, %q)", test.str, test.pat))
		succ, start, end, _ := Find(test.str, test.pat, false)
		if succ != test.succ {
			t.Errorf("find('%s', '%s') returned %t, expected %t", test.str, test.pat, succ, test.succ)
			return
		}
		if succ {
			substr := test.str[start:end]
			if substr != test.cap {
				t.Errorf("find('%s', '%s') => substr '%s' does not match expected '%s' (got %d and %d as start/end)", test.str, test.pat, substr, test.cap, start, end)
				return
			}
		}
	}
}
