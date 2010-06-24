package luapatterns

import (
	"bytes"
	"fmt"
)

const (
	CAP_UNFINISHED = -1
	CAP_POSITION = -2
	L_ESC = "%"
	SPECIALS = "^$*+?.([%-"
	LUA_MAXCAPTURES = 32	// arbitrary
)

// Returns the index in 's1' where the 's2' can be found, or -1
func lmemfind(s1 []byte, s2 []byte) int {
	fmt.Printf("Begin lmemfind('%s', '%s')\n", s1, s2)
	l1, l2 := len(s1), len(s2)
	if l2 == 0 {
		return 0
	} else if l2 > l1 {
		return -1
	} else {
		init := bytes.IndexByte(s1, s2[0])
		end := init + l2
		for end <= l1 && init != -1 {
			//fmt.Printf("l1: %d, l2: %d, init: %d, end: %d, slice: %s\n", l1, l2, init, end, s1[init:end])
			init++		// 1st char is already checked by IndexBytes
			if bytes.Equal(s1[init - 1:end], s2) {
				return init - 1
			} else {	// find the next 'init' and try again
				next := bytes.IndexByte(s1[init:], s2[0])
				if next == -1 {
					return -1
				} else {
					init = init + next
					end = init + l2
				}
			}
		}
	}

	return -1
}

func str_find_aux(s, p []byte, init int, plain bool) (bool, int, int) {
	l1, l2 := len(s), len(p)

	if init < 0 {
		init = 0
	} else if init > l1 {
		init = l1
	}

	// check if we can do a plain search
	if plain || bytes.IndexAny(p, SPECIALS) == -1 {
		if index := lmemfind(s[init:], p); index != -1 {
			return true, index, index + l2
		}
	} else {
		//ms := new(MatchState)

		// Initialize tha match state
		// do 
		// 		if res = match(ms, s1, p) != NULL
		//			if find then push start, end of match, and captures
		//			else just push captures
		// while <condition>
		// 		anchor is not 1
		// 		s1++ < ms.src_end

		// return nil, nothing found
	}

	return false, -1, -1
}
