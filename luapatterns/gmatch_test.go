package luapatterns

import (
	"fmt"
	"reflect"
	"testing"
)

type GmatchTest struct {
	str  string
	pat  string
	vals [][]string
}

var GmatchTests = []GmatchTest{
	GmatchTest{"Apple", "%w", [][]string{
		[]string{"A"},
		[]string{"p"},
		[]string{"p"},
		[]string{"l"},
		[]string{"e"},
	}},
	GmatchTest{"Apple", "(%w)(%w)", [][]string{
		[]string{"A", "p"},
		[]string{"p", "l"},
	}},
}

func TestGmatch(t *testing.T) {
	enableDebug = false
	for _, test := range GmatchTests {
		debug(fmt.Sprintf("==== %s ====", test))

		idx := 0
		for caps := range Gmatch(test.str, test.pat) {
			if !reflect.DeepEqual(caps, test.vals[idx]) {
				t.Errorf("Captures do not match: got %s espected %s", caps, test.vals)
			}
			idx = idx + 1
		}
	}
}
