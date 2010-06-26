package luapatterns

import (
	"fmt"
	"reflect"
	"testing"
)

// stopping errors
var foo = fmt.Sprintf("blah")

func get_onecapture(ms *matchState, i int, s, e *sptr) []byte {
	if i >= ms.level {
		if i == 0 {		// ms->level == 0 too
			//fmt.Printf("Returning whole string\n")

			return s.getStringLen(e.length() - s.length())
		} else {
			panic("invalid capture index")
		}
	} else {
		var l int = ms.capture[i].len
		if l == CAP_UNFINISHED {
			panic("unfinished capture")
		}
		if l == CAP_POSITION {
			// TODO: Find a way to fix this
			panic("position captures not supported")
		} else {
			return ms.capture[i].init.getStringLen(l)
		}
	}
	panic("never reached")
}

func matchWrapper(s, p string) (bool, []string) {
	// Create string pointers
	sp := &sptr{[]byte(s), 0}
	pp := &sptr{[]byte(p), 0}

	// Create a new matchstate
	ms := new(matchState)
	ms.src_init = sp
	ms.src_end = sp.cloneAt(5)

	s1 := ms.src_init.clone()

	for {
		var res *sptr = match(ms, s1, pp)
		if res != nil {
			// fmt.Printf("===== GOT MATCH =====\n")
			// fmt.Printf("ms.level: %d\n", ms.level)
			// fmt.Printf("Res: %s\n", res)
			// fmt.Printf("s1: %s\n", s1)
			// fmt.Printf("Captures: \n")
			for i := 0; i < LUA_MAXCAPTURES; i++ {
				capt := ms.capture[i]
				if capt.init != nil {
					//fmt.Printf("Index %d: str: %s, index: %d, len: %d\n", i, capt.init.str, capt.init.index, capt.len)
				}
			}

			// Fetch the captures
			captures := new([LUA_MAXCAPTURES]string)

			var i int
			var nlevels int
			if ms.level == 0 && s1 != nil {
				nlevels = 1
			} else {
				nlevels = ms.level
			}

			for i = 0; i < nlevels; i++ {
				captures[i] = string(get_onecapture(ms, i, s1, res))
				//fmt.Printf("CAP[%d] = %s\n", i, captures[i])
			}

			return true, captures[0:nlevels]
		}
		if s1.postInc(1) >= ms.src_end.index {
			return false, nil
		}
	}
	panic("never reached")
}

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
}

func TestMatch(t *testing.T) {
	for _, test := range MatchTests {
		succ, caps := matchWrapper(test.str, test.pat)
		if succ != test.succ {
			t.Errorf("match('%s', '%s') returned %t instead of expected %t", test.str, test.pat, succ, test.succ)
		}
		if !reflect.DeepEqual(caps, test.caps) {
			t.Errorf("Captures do not match: got %s expected %s", caps, test.caps)
		}
	}
}
