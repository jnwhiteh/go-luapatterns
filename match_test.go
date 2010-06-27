package luapatterns

import (
	"fmt"
	"reflect"
	"testing"
)

// stopping errors
var foo = fmt.Sprintf("blah")

type MatchTest struct {
	str string
	pat string
	succ bool
	caps []string
}

var MatchTests = []MatchTest{
	MatchTest{"Apple", "[Aa]pple", true, []string{"Apple"}},
	MatchTest{"Apple", "apple", false, []string{}},
	MatchTest{"Apple", "(Ap)ple", true, []string{"Ap"}},
	MatchTest{"Apple", "(Ap)p(le)", true, []string{"Ap", "le"}},
	MatchTest{"Apple", "A(pp)(le)", true, []string{"pp", "le"}},
	MatchTest{"apple", "a[Pp][Pp]le", true, []string{"apple"}},
	MatchTest{"a1\x00a1aaA>a aAa1a ag\x00a", "%a%A%c%C%d%D%l%L%p%P%s%S%u%U%w%w%W%x%X%z%Z",
		true, []string{"a1\x00a1aaA>a aAa1a ag\x00a"}},
	MatchTest{"1", "%a", false, nil},
	MatchTest{"a", "%c", false, nil},
	MatchTest{"a", "%d", false, nil},
	MatchTest{"A", "%l", false, nil},
	MatchTest{"a", "%p", false, nil},
	MatchTest{"a", "%s", false, nil},
	MatchTest{"a", "%u", false, nil},
}

func TestMatch(t *testing.T) {
	for _, test := range MatchTests {
		debug(fmt.Sprintf("==== %s ====", test))
		succ, caps := MatchString(test.str, test.pat, 0)
		if succ != test.succ {
			t.Errorf("match('%s', '%s') returned %t instead of expected %t", test.str, test.pat, succ, test.succ)
		} else if !reflect.DeepEqual(caps, test.caps) {
			t.Errorf("Captures do not match: got %s expected %s", caps, test.caps)
		}
	}
}
