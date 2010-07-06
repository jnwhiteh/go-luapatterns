package luapatterns

import (
	"bytes"
	"io"
	"strings"
	"os"
)

var enableDebug bool = false
func debug(s string) {
	if enableDebug {
		io.WriteString(os.Stderr, s + "\n")
	}
}

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
	debug("check_capture")
	l = l - int('1')		// change capture index to offset in capture array
	if l < 0 || l >= ms.level || ms.capture[l].len == CAP_UNFINISHED {
		panic("invalid capture index")
	}
	return l
}

func capture_to_close(ms *matchState) int {
	debug("capture_to_close")
	level := ms.level
	for level--; level >=0; level-- {
		if ms.capture[level].len == CAP_UNFINISHED {
			return level
		}
	}
	panic("invalid pattern capture")
}

func classend(ms *matchState, pp *sptr) *sptr {
	debug("classend")
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
			for {							// look for an ']'
				if p.getChar() == 0 {
					panic("malformed pattern (missing ']')")
				}
				pch := p.getChar()
				p.postInc(1)
				if pch == L_ESC && p.getChar() != 0 {
					p.postInc(1)
				}
				// while condition at the end
				if p.getChar() == ']' {
					break
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
	debug("match_class")
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
	debug("matchbracketclass")
	p := pp.clone()
	var sig bool = true
	if p.getCharAt(1) == '^' {
		sig = false
		p.postInc(1)		// skip the '^'
	}
	for p.preInc(1) < ec.index {
		if p.getChar() == L_ESC {
			p.postInc(1)
			if match_class(c, p.getChar()) {
				return sig
			}
		} else if (p.getCharAt(1) == '-') && p.index + 2 < ec.index {
			p.postInc(2)
			if p.getCharAt(-2) <= c && c <= p.getChar() {
				return sig
			}
		} else if p.getChar() == c {
			return sig
		}
	}

	return !sig
}

func singlematch(c byte, pp, epp *sptr) bool {
	debug("singlematch")
	// clone pointers that get pass outside this function
	p, ep := pp.clone(), epp.clone()
	switch p.getChar() {
		case '.': return true
		case L_ESC: return match_class(c, p.getCharAt(1))
		case '[': return matchbracketclass(c, p, ep.cloneAt(-1))
		default: return p.getChar() == c
	}

	return false
}

func matchbalance(ms *matchState, sp, p *sptr) *sptr {
	debug("matchbalance")
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
	debug("max_expand")
	// clone pointers that get pass outside this function
	s, p, ep := sp.clone(), pp.clone(), epp.clone()

	var i int = 0		// count maximum expand for item
	for s.index + i < ms.src_end.index && singlematch(s.getCharAt(i), p, ep) {
		i++
	}

	// keeps trying to match with the maximum repititions
	for i >= 0 {

		res := match(ms, s.cloneAt(i), ep.cloneAt(1))
		if res != nil {
			return res
		}
		i--				// else didn't match; reduce 1 repetition to try again 
	}

	return nil
}

func min_expand(ms *matchState, sp, pp, epp *sptr) *sptr {
	debug("min_expand")
	// clone pointers that get pass outside this function
	s, p, ep := sp.clone(), pp.clone(), epp.clone()

	for {
		res := match(ms, s, ep.cloneAt(1))
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
	debug("start_capture")
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
	if res = match(ms, s, p); res == nil {		// match failed?
		ms.level--								// undo capture
	}
	return res
}

func end_capture(ms *matchState, sp, pp *sptr) *sptr {
	debug("end_capture")
	s, p := sp.clone(), pp.clone()
	var l int = capture_to_close(ms)
	var res *sptr
	ms.capture[l].len = s.index - ms.capture[l].init.index		// close capture
	if res = match(ms, s, p); res == nil {						// match failed?
		ms.capture[l].len = CAP_UNFINISHED						// undo capture
	}
	return res
}

func match_capture(ms *matchState, sp *sptr, l int) *sptr {
	debug("match_capture")
	s := sp.clone()

	l = check_capture(ms, l)
	lenc := ms.capture[l].len
	capstr := ms.capture[l].init.getStringLen(lenc)
	sstr := s.getStringLen(lenc)

	if ms.src_end.index - s.index >= lenc && bytes.Compare(capstr, sstr) == 0 {
		return s.cloneAt(lenc)
	} else {
		return nil
	}

	panic("never reached")
}

func match(ms *matchState, sp, pp *sptr) *sptr {
	debug("match")
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
			debug("dflt label")
			var ep *sptr = classend(ms, p)		// points to what is next
			var m bool = s.index < ms.src_end.index && singlematch(s.getChar(), p, ep)
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

func get_onecapture(ms *matchState, i int, s, e *sptr) []byte {
	debug("get_onecapture")
	if i >= ms.level {
		if i == 0 {		// ms->level == 0 too
			// return whole match
			return s.getStringLen(e.index - s.index)
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

func find_and_capture(s, p []byte) (bool, int, int, [][]byte) {
	slen := len(s)

	// Turn s and p into string pointers
	sp := &sptr{s, 0}
	pp := &sptr{p, 0}

	ms := new(matchState)
	ms.src_init = sp
	ms.src_end = sp.cloneAt(slen)
	s1 := ms.src_init.clone()

	var anchor bool
	if pp.getChar() == '^' {
		pp.postInc(1)
		anchor = true
	} else {
		anchor = false
	}

	for {
		var res *sptr = match(ms, s1, pp)
		if res != nil {
			start := s1.index - sp.index
			end := res.index - sp.index

			// Fetch the captures
			captures := new([LUA_MAXCAPTURES][]byte)

			var i int
			var nlevels int
			if ms.level == 0 && s1 != nil {
				nlevels = 1
			} else {
				nlevels = ms.level
			}

			for i = 0; i < nlevels; i++ {
				captures[i] = get_onecapture(ms, i, s1, res)
			}

			return true, start, end, captures[0:nlevels]
		}
		if s1.postInc(1) >= ms.src_end.index || anchor {
			return false, -1, -1, nil
		}
	}

	panic("never reached")
}

// Looks for the first match of pattern p in the string s. If it finds one,
// then match returns true and the captures from the pattern; otherwise it
// returns false, nil.  If pattern specifies no captures, then the whole match
// is returned.
func Match(s, p  string) (bool, []string) {
	succ, _, _, caps := find_and_capture([]byte(s), []byte(p))
	scaps := make([]string, LUA_MAXCAPTURES)
	for idx, str := range caps {
		scaps[idx] = string(str)
	}
	return succ, scaps[0:len(caps)]
}

// Same as the Match function, however operates directly on byte arrays rather
// than strings. This package operates natively in bytes, so this function is
// called by Match to perform it's work. 
func MatchBytes(s, p []byte) (bool, [][]byte) {
	succ, _, _, caps := find_and_capture(s, p)
	return succ, caps
}

// Looks for the first match of pattern p in the string s. If it finds a match,
// then find returns the indices of s where this occurrence starts and ends;
// otherwise, it returns nil. If the pattern has captures, they are returned in
// an array. If the argument 'plain' is set to 'true', then this function
// performs a plain 'find substring' operation with no characters in the
// pattern being considered magic.
//
// Note that the indices returned from this function will NOT match the
// versions returned by the equivalent Lua string and pattern due to the
// differences in slice semantics and array indexing. 
//
// You can rely on the fact that s[startIdx:endIdx] will be the entire portion
// of the string that matched the pattern.
func Find(s, p string, plain bool) (bool, int, int, []string) {
	sb, pb := []byte(s), []byte(p)
	succ, start, end, caps := FindBytes(sb, pb, plain)

	scaps := make([]string, LUA_MAXCAPTURES)
	for idx, str := range caps {
		scaps[idx] = string(str)
	}

	return succ, start, end, scaps[0:len(caps)]
}

// Same as the Find function, however operates directly on byte arrays rather
// than strings. This package operates natively in bytes, so this function is
// called by Find to perform it's work. 
func FindBytes(s, p []byte, plain bool) (bool, int, int, [][]byte) {
	if plain || bytes.IndexAny(p, SPECIALS) == -1 {
		if index := lmemfind(s, p); index != -1 {
			return true, index, index + len(p), nil
		} else {
			return false, -1, -1, nil
		}
	}

	// Do a normal match with captures
	return find_and_capture(s, p)
}

func add_s(ms *matchState, b *bytes.Buffer, s, e *sptr, news []byte) {
	l := len(news)
	for i := 0; i < l; i++ {
		if news[i] != L_ESC {
			b.WriteByte(news[i])
		} else {
			i++						// skip ESC (%)
			if (!isdigit(news[i])) {
				b.WriteByte(news[i])
			} else if news[i] == '0' {
				b.Write(s.getStringLen(e.index - s.index))
			} else {
				cidx := int(news[i] - '1')
				b.Write(get_onecapture(ms, cidx, s, e))
			}
		}
	}
}

// Replaces up-to n instances of patt with repl in the source string. In the
// string repl, the charachter % works as an escape cahracter: any sequence of
// the form %n, with n between 1 and 9, stands for the value of the n-th
// captured substring due to the match with patt. The sequence %0 stands for
// the entire match. The sequence %% stands for a single % in the resulting
// string. A value of -1 for n will replace all instances of patt with repl.
func Replace(src, patt, repl string, max int) (string, int) {
	res, n := ReplaceBytes([]byte(src), []byte(patt), []byte(repl), max)
	return string(res), n
}

// Same as the Replace function, however operates directly on byte arrays
// rather than strings. This package operates natively in bytes, so this
// function is called by Replace to perform it's work. 
func ReplaceBytes(src, patt, repl []byte, max int) ([]byte, int) {
	srcl := len(src)
	pattp := &sptr{patt, 0}
	srcp := &sptr{src, 0}

	var anchor bool = false

	if pattp.getChar() == '^' {
		anchor = true
		pattp.postInc(1)
	}

	var n int = 0
	var b bytes.Buffer
	ms := new(matchState)
	ms.src_init = srcp
	ms.src_end = srcp.cloneAt(srcl)
	for n < max || max == -1 {
		ms.level = 0
		e := match(ms, srcp, pattp)
		if e != nil {
			n++
			add_s(ms, &b, srcp, e, repl)			// Use add_s directly here
		}
		if e != nil && e.index > srcp.index {	// Non empty match
			srcp.index = e.index				// skip it
		} else if srcp.index < ms.src_end.index {
			b.WriteByte(srcp.getChar())
			srcp.postInc(1)
		} else {
			break
		}

		if anchor {
			break
		}
	}
	b.Write(srcp.getStringLen(ms.src_end.index - srcp.index))
	return b.Bytes(), n
}
