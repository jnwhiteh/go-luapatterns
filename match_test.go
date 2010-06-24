package luapatterns

import (
	"testing"
	"fmt"
)

func get_onecapture(ms *matchState, i int, s, e *sptr) []byte {
	if i >= ms.level {
		if i == 0 {		// ms->level == 0 too
			fmt.Printf("Returning whole string\n")

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
			fmt.Printf("===== GOT MATCH =====\n")
			fmt.Printf("ms.level: %d\n", ms.level)
			fmt.Printf("Res: %s\n", res)
			fmt.Printf("s1: %s\n", s1)
			fmt.Printf("Captures: \n")
			for i := 0; i < LUA_MAXCAPTURES; i++ {
				capt := ms.capture[i]
				if capt.init != nil {
					fmt.Printf("Index %d: str: %s, index: %d, len: %d\n", i, capt.init.str, capt.init.index, capt.len)
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
			}

			return true, captures[0:nlevels]
		}
		if s1.postInc(1) >= ms.src_end.index {
			return false, nil
		}
	}
	panic("never reached")
}

func TestMatch(t *testing.T) {
	succ, caps := matchWrapper("Apple", "[Aa]p(p)le")
	t.Errorf("succ: %t", succ)
	t.Errorf("caps: %s", caps)
}

func _TestMatch(t *testing.T) {
	ms := new(matchState)
	str := []byte("Apple")

	ms.src_init = &sptr{str, 0}
	ms.src_end = &sptr{str, 5}

	s1 := ms.src_init.clone()

	p := &sptr{[]byte("([Aa])pple"), 0}

	for {
		var res *sptr = match(ms, s1, p)
		if res != nil {
			t.Errorf("MATCH FOUND: %s", ms)
		}
		if s1.postInc(1) >= ms.src_end.index {
			t.Errorf("Breaking out:")
			t.Errorf("res: %s", res)
			t.Errorf("s1: %s", s1)
			t.Errorf("p: %s", p)
			t.Errorf("ms.src_end: %s", ms.src_end)
			t.Errorf("ms.src_init: %s", ms.src_init)
			t.Errorf("ms.level: %d", ms.level)
			t.Errorf("len(ms.capture): %d cap, %d", len(ms.capture), cap(ms.capture))
			break
		}
	}

	t.Errorf("Failed to find a match")
}
