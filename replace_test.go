package luapatterns

import (
	"testing"
)

type replaceTest struct {
	src string
	pat string
	rep string
	max int
	res string
	n int
}

var replaceTests = []replaceTest{
	replaceTest{"\243lo \243lo", "\243", "x", -1, "xlo xlo", 2},		// 243 = ú
	replaceTest{"alo \243lo  ", " +$", "", -1, "alo \243lo", 1},		// trim
	replaceTest{"  alo alo  ", "^%s*(.-)%s*$", "%1", -1, "alo alo", 1},	// double trim
	// POSITION CAPTURES NOT SUPPORTED
	replaceTest{"abc=xyz", "(%w*)(%p)(%w+)", "%3%2%1", -1, "xyz=abc", 1},
	replaceTest{"aei", "$", "\x00ou", -1, "aei\x00ou", 1},
	replaceTest{"", "^", "r", -1, "r", 1},
	replaceTest{"", "$", "r", -1, "r", 1},
}

func TestReplace(t *testing.T) {
	enableDebug = false

	for _, test := range replaceTests {
		res, n := Replace(test.src, test.pat, test.rep, test.max)
		if res != test.res {
			t.Errorf("replace('%s', '%s', '%s', %d) returned '%s', expected '%s'",
				test.src, test.pat, test.rep, test.max, res, test.res)
		} else if n != test.n {
			t.Errorf("replace('%s', '%s', '%s', %d) performed %d replacements, expected %d", 
				test.src, test.pat, test.rep, test.max, n, test.n)
		}
	}
}
