package luapatterns

import (
	"bytes"
	"strings"
	"fmt"
)

const (
	LUA_MAXCAPTURES = 32	// arbitrary

	CAP_UNFINISHED = -1
	CAP_POSITION = -2

	L_ESC = '%'
	SPECIALS = "^$*+?.([%-"
)

type capture struct {
	init *sptr
	len int
}

type matchState struct {
	src_init *sptr
	src_end *sptr
	level int
	capture [LUA_MAXCAPTURES]capture
}

func check_capture(ms *matchState, l int) int {
	// TODO: Why the hell is this being done?
	var one byte = '1'
	l = l - int(one)
	if l < 0 || l >= ms.level || ms.capture[l].len == CAP_UNFINISHED {
		panic("invalid capture index")
	}
	return l
}

func capture_to_close(ms *matchState) int {
	level := ms.level
	for level--; level >=0; level-- {
		if ms.capture[level].len == CAP_UNFINISHED {
			return level
		}
	}
	panic("invalid pattern capture")
}

func classend(ms *matchState, pp *sptr) *sptr {
	p := pp.clone()
	char := p.getChar()
	p.postInc(1)
	switch (char) {
		case L_ESC: {
			if p.getChar() == 0 {
				panic("malformed pattern (ends with '%'")
			}
			p.preInc(1)
			return p
		}
		case '[': {
			if p.getChar() == '^' {
				p.preInc(1)
			}
			for p.getChar() != ']' {		// look for an ']'
				if p.getChar() == 0 {
					panic("malformed pattern (missing ']')")
				}
				if p.postInc(1); p.getChar() == L_ESC && p.getChar() != 0 {
					p.preInc(1)				// skip escapes (e.g. '%]')
				}
			}
			p.preInc(1)
			return p
		}
		default: {
			return p
		}
	}
	panic("never reached")
}

func match_class(c byte, cl byte) bool {
	var res bool

	cllower := strings.ToLower(string(cl))[0]
	switch cllower {
		case 'a': res = isalpha(c)
		case 'c': res = iscntrl(c)
		case 'd': res = isdigit(c)
		case 'l': res = islower(c)
		case 'p': res = ispunct(c)
		case 's': res = isspace(c)
		case 'u': res = isupper(c)
		case 'w': res = isalnum(c)
		case 'x': res = isxdigit(c)
		case 'z': res = (c == 0)
		default: return cl == c
	}

	if islower(cl) {
		return res
	}

	return !res
}

func matchbracketclass(c byte, pp, ec *sptr) bool {
	p := pp.clone()
	var sig bool = true
	if p.getCharAt(1) == '^' {
		sig = false
		p.postInc(1)		// skip the '^'
	}
	for {
		p.preInc(1)
		if p.index < ec.index {
			p.postInc(1)
			if match_class(c, p.getChar()) {
				return sig
			}
		} else if p.getCharAt(1) == '-' && p.index + 2 < ec.index {
			return sig
		} else if p.getChar() == uint8(c) {
			return sig
		}
	}
	return !sig
}

func singlematch(c byte, pp, epp *sptr) bool {
	// clone pointers that get pass outside this function
	p, ep := pp.clone(), epp.clone()
	fmt.Printf("singlematch on '%c'\n", p.getChar())
	switch p.getChar() {
		case '.': return true
		case L_ESC: return match_class(c, p.getCharAt(1))
		case '[': return matchbracketclass(c, p, ep.cloneAt(-1))
		default: return p.getChar() == c
	}

	return false
}

func matchbalance(ms *matchState, sp, p *sptr) *sptr {
	s := sp.clone()
	if p.getChar() == 0 || p.getCharAt(1) == 0 {
		panic("unbalanced pattern")
	}
	if s.getChar() != p.getChar() {
		return nil
	} else {
		var b byte = p.getChar()
		var e byte = p.getCharAt(1)
		var cont int = 1
		for s.preInc(1) < ms.src_end.index {
			if s.getChar() == e {
				cont = cont - 1
				if cont == 0 {
					s.preInc(1)
					return s
				}
			} else if s.getChar() == b {
				cont++
			}
		}
	}
	return nil		// string ends out of balance
}

func max_expand(ms *matchState, sp, pp, epp *sptr) *sptr {
	// clone pointers that get pass outside this function
	s, p, ep := sp.clone(), pp.clone(), epp.clone()

	var i int = 0		// count maximum expand for item
	for s.index + i < ms.src_end.index && singlematch(s.getCharAt(i), p, ep) {
		i++
	}
	// keeps trying to match with the maximum repititions
	for i >= 0 {
		snext := s.clone()
		epnext := ep.clone()
		snext.preInc(i)
		epnext.preInc(1)
		var res *sptr = match(ms, s, ep)
		if res != nil {
			return res
		}
		i--				// else didn't match; reduce 1 repetition to try again
	}
	return nil
}

func min_expand(ms *matchState, sp, pp, epp *sptr) *sptr {
	// clone pointers that get pass outside this function
	s, p, ep := sp.clone(), pp.clone(), epp.clone()

	for {
		ep.preInc(1)
		res := match(ms, s, ep)
		if res != nil {
			return res
		} else if s.index < ms.src_end.index && singlematch(s.getChar(), p, ep) {
			s.postInc(1)		// try with one more repetition
		} else {
			return nil
		}
	}

	panic("never reached")
}

func start_capture(ms *matchState, sp, pp *sptr, what int) *sptr {
	// clone pointers that get pass outside this function
	s, p := sp.clone(), pp.clone()

	var res *sptr
	var level int = ms.level
	if level >= LUA_MAXCAPTURES {
		panic("too many captures")
	}
	ms.capture[level].init = s
	ms.capture[level].len = what
	ms.level = level + 1
	if res = match(ms, s, p); res != nil {		// match failed?
		ms.level--								// undo capture
	}
	return res
}

func end_capture(ms *matchState, sp, pp *sptr) *sptr {
	s, p := sp.clone(), pp.clone()
	var l int = capture_to_close(ms)
	var res *sptr
	ms.capture[l].len = s.index - ms.capture[l].init.index		// close capture
	if res = match(ms, s, p); res == nil {						// match failed?
		ms.capture[l].len = CAP_UNFINISHED						// undo capture
	}
	return res
}

// TODO: Is this function correct? Had to do a bunch of translation
func match_capture(ms *matchState, sp *sptr, l int) *sptr {
	s := sp.clone()
	var length int
	l = check_capture(ms, l)
	length = ms.capture[l].len
	capstr := ms.capture[l].init.str[ms.capture[l].init.index:][1:length]
	sstr := s.str[1:length]

	if ms.src_end.index - s.index >= length && bytes.Compare(capstr, sstr) == 0 {
		s.preInc(length)
		return s
	}
	return nil
}

func match(ms *matchState, sp, pp *sptr) *sptr {
	s, p := sp.clone(), pp.clone()

	init:						// use goto's to optimize tail recursion
	switch(p.getChar()) {
		case '(': {							// start capture
			if p.getCharAt(1) == ')' {		// position capture
				return start_capture(ms, s, p.cloneAt(2), CAP_POSITION)
			} else {
				return start_capture(ms, s, p.cloneAt(1), CAP_UNFINISHED)
			}
		}
		case ')': {							// end capture
			p.preInc(1)
			return end_capture(ms, s, p)
		}
		case L_ESC: {
			switch p.getCharAt(1) {
				case 'b': {					// balanced string?
					s = matchbalance(ms, s, p.cloneAt(2))
					if s == nil {
						return nil
					}
					p.preInc(4)
					goto init				// else return match(ms, s, p+4)
				}
				case 'f': {					// frontier
					var ep *sptr
					var previous byte
					p.preInc(2)
					if p.getChar() != '[' {
						panic("Missing '[' after '%f' in pattern")
					}
					ep = classend(ms, p)
					if s.index == ms.src_init.index {
						previous = 0
					} else {
						previous = s.getCharAt(-1)
					}
					if matchbracketclass(previous, p, ep.cloneAt(-1)) ||
						!matchbracketclass(s.getChar(), p, ep.cloneAt(-1)) {
							return nil
					}
					p = ep; goto init		// else return match(ms, s, ep)
				}
				default: {
					if isdigit(p.getCharAt(1)) {	// capture results (%0-%9)?
						s = match_capture(ms, s, int(p.getCharAt(1)))
						if s == nil {
							return nil
						}
						p.preInc(2); goto init		// else return match(ms, s, p+2)
					}
					goto dflt						// case default
				}
			}
		}
		case 0: {	// end of pattern
			return s	// match succeeded
		}
		case '$': {
			if p.getCharAt(1) == 0 {				// is the '$' the last char in pattern?
				if s.index == ms.src_end.index {	// check end of string
					return s
				} else {
					return nil
				}
			} else {
				goto dflt
			}
		}
		default:		// it is a pattern item
			dflt:
			fmt.Printf("In dflt\n")
			var ep *sptr = classend(ms, p)		// points to what is next
			fmt.Printf("sptr: %s\n", ep)
			var m bool = s.index < ms.src_end.index && singlematch(s.getChar(), p, ep)
			fmt.Printf("cond: %t\n", s.index < ms.src_end.index)
			fmt.Printf("m: %t\n", m)
			switch ep.getChar() {
				case '?': {		// optional
					var res *sptr
					if m {
						res = match(ms, s.cloneAt(1), ep.cloneAt(1))
						if res != nil {
							return res
						}
					}
					p = ep.cloneAt(1)
					goto init				// else return match(ms, s, ep+1)
				}
				case '*': {		// 0 or more repetitions
					return max_expand(ms, s, p, ep)
				}
				case '+': {		// 1 or more repetitions
					if m {
						return max_expand(ms, s.cloneAt(1), p, ep)
					} else {
						return nil
					}
				}
				case '-': {		// 0 or more repetitions (minimum)
					return min_expand(ms, s, p, ep)
				}
				default: {
					if !m {
						return nil
					} else {
						s.preInc(1)
						p = ep
						goto init	// else return match(ms, s+1, ep)
					}
				}
			}
	}
	panic("never reached")
}

// Returns the index in 's1' where the 's2' can be found, or -1
func lmemfind(s1 []byte, s2 []byte) int {
	//fmt.Printf("Begin lmemfind('%s', '%s')\n", s1, s2)
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
