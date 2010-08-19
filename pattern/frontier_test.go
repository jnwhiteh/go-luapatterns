package luapatterns

import (
	"testing"
)

type frontierTest struct {
	src string
	pat string
	rep string
	res string
}

var frontierTests = []frontierTest{
	frontierTest{"aaa aa a aaa a", "%f[%w]a", "x", "xaa xa x xaa x"},
	frontierTest{"[[]] [][] [[[[", "%f[[].", "x", "x[]] x]x] x[[["},
	frontierTest{"01abc45de3", "%f[%d]", ".", ".01abc.45de.3"},
	frontierTest{"function", "%f[\x01-\xff]%w", ".", ".unction"},
	frontierTest{"function", "%f[^\x01-\xff]", ".", "function."},
}

// TODO: Implement frontier pattern
func DisabledTestFrontier(t *testing.T) {
	enableDebug = true

	for _, test := range frontierTests {
		res, _ := Replace(test.src, test.pat, test.rep, -1)
		if res != test.res {
			t.Errorf("replace('%s', '%s', '%s', %d) returned '%s', expected '%s'",
				test.src, test.pat, test.rep, -1, res, test.res)
			return
		}
	}
}
