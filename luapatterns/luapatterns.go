package luapatterns

import "bytes"
import "fmt"

var _fmtfix string = fmt.Sprintf("fmtfix")
var enableDebug bool = false

func debug(str string) {
	if enableDebug {
		fmt.Printf(str)
		fmt.Printf("\n")
	}
}

// We can't INDEX outside the bounds of an array, but we can make a slice at a
// higher index, so long as it doesn't have an end index as well. So, with:
//
// var str []byte = "hello"
// str[5] is illegal but
// str[7:] is legal
//
// Trace of match("Apple", "%w+")
// 	- L_ESC branch
// 	- go to dflt branch due to not being a capture reference
//	- classend(ms, p)
//  	- ep  = p[2:] (just moves p to the +)
//	- m gets set to true if the next character if s would match the pattern item
//	- check for ep (which is the + in this case)
// 	- max_expands(ms, s, p, ep)
//    - i = how many matches we could possibly make
//	  - Try to match with the remainder and backtrack one reptition if not possible
//	- match(ms, s[i+1:], ep[1:])
//	  - This is the string after the max matches, and after the + in pattern
// 	- pattern is empty, so match succeeds by returning 's'. Why? I don't know.

// There is a wrapper that moves through the string trying to match portions of
// it, that's how things end up working. See find_and_capture in
// luapatterns.go, which just moves through the string incremending as we move
// through the string. Perhaps its best to start there writing that.

// Rather than passing ep as a pointer, we should pass it as a numeric index
// into p, this way we can easily step back and forth as needed. This should
// make most of the implementation straightforward.

const (
	LUA_MAXCAPTURES = 32
	CAP_UNFINISHED  = -1
	SPECIALS        = "^$*+?.([%-"
)

type capture struct {
	src []byte
	len int
}

type matchState struct {
	src     []byte
	level   int
	capture []capture
}

func (ms matchState) String() string {
	return "<ms>"
	result := fmt.Sprintf("matchState{'%s', %d, ", ms.src, ms.level)
	if len(ms.capture) == 0 {
		result = result + "{}}"
	} else {
		result = result + fmt.Sprintf("%q}", ms.capture)
	}

	return result
}

// func (ms matchState) String() string {
// 	return "<ms>"
// }

func tolower(c byte) byte {
	if c >= 65 && c <= 90 { // upper case
		return c + 32
	}
	return c
}

// Returns whether or not a byte matches a specified character class, which may
// simply be a non-special character itself.

func match_class(c byte, cl byte) (res bool) {
	//debug(fmt.Sprintf("match_class(%q, %q)", c, cl))

	cllower := tolower(cl)
	switch cllower {
	case 'a':
		res = isalpha(c)
	case 'c':
		res = iscntrl(c)
	case 'd':
		res = isdigit(c)
	case 'l':
		res = islower(c)
	case 'p':
		res = ispunct(c)
	case 's':
		res = isspace(c)
	case 'u':
		res = isupper(c)
	case 'w':
		res = isalnum(c)
	case 'x':
		res = isxdigit(c)
	case 'z':
		res = (c == 0)
	default:
		return cl == c
	}

	if islower(cl) {
		return res
	}

	return !res // handle upper-case reverse classes
}

// Returns whether or not a given character matches the character class
// specified in the pattern.

func matchbracketclass(c byte, p []byte, ec []byte) bool {
	//debug(fmt.Sprintf("matchbracketclass('%c', %q, %q)", c, p, ec))
	var sig bool = true
	if p[1] == '^' {
		sig = false
		p = p[1:]
	}
	for p = p[1:]; len(p) > len(ec); p = p[1:] {
		if p[0] == '%' {
			p = p[1:]
			if match_class(c, p[0]) {
				return sig
			}
		} else if p[1] == '-' && (len(p)-2 > len(ec)) {
			if p[0] <= c && c <= p[2] {
				return sig
			}
		} else if p[0] == c {
			return sig
		}
	}

	return !sig
}

// Returns whether or not a single character matches the pattern currently
// being examined. It needs to 'lookahead' in order to accomplish this, for
// example when the pattern is a character class like %l. This is the purpose
// of the argument ep.

func singlematch(c byte, p []byte, ep []byte) bool {
	//debug(fmt.Sprintf("singlematch('%c', %q, %q)", c, p, ep))
	switch p[0] {
	case '.':
		return true
	case '%':
		return match_class(c, p[1])
	case '[':
		{
			// Move ep back a character
			ep = p[len(p)-len(ep)-1:]
			return matchbracketclass(c, p, ep)
		}
	default:
		return p[0] == c
	}

	return false
}

// Returns the portion of the source string that matches the balance pattern
// specified, where b is the start and e is the end of the balance pattern.

func matchbalance(ms *matchState, s, p []byte) []byte {
	//debug(fmt.Sprintf("matchbalance(%q, %q, %q)", ms, s, p))
	if len(p) <= 1 {
		// error: unbalanced pattern
		return nil
	}
	if s[0] != p[0] {
		return nil
	} else {
		var b byte = p[0]
		var e byte = p[1]
		var cont int = 1

		// ms.src_end in the original C source is a pointer to the end of the
		// source string (whatever that means specifically). This loop wants to
		// ensure that s remains less than this pointer. Since we're not
		// dealing with pointers, we should be able to just run the loop until
		// s runs out.

		for s = s[1:]; len(s) > 0; s = s[1:] {
			if s[0] == e {
				cont = cont - 1
				if cont == 0 {
					return s[1:]
				}
			} else if s[0] == b {
				cont = cont + 1
			}
		}
	}

	// error: strings ends out of balance
	return nil
}

// Return the maximum portion of the source string that matches the given
// pattern (equates to the '+' or '*' operator)

func max_expand(ms *matchState, s, p, ep []byte) []byte {
	//debug(fmt.Sprintf("max_expand(%q, %q, %q, %q)", ms, s, p, ep))

	// Run through the string to find the maximum number of matches that are
	// possible for the pattern item.

	var i int = 0 // count maximum expand for item
	for i = 0; i < len(s) && singlematch(s[i], p, ep); i++ {
	}

	//debug(fmt.Sprintf("Can match up to %d times", i))

	// Try to match with maximum reptitions
	for i >= 0 {
		res := match(ms, s[i:], ep[1:])
		if res != nil {
			return res
		} else {
			// Reduce 1 repetition and try again
			i--
		}
	}
	return nil
}

// Returns the minimum portion of the source string that matches the given
// pattern (equates to the '-' operator)

func min_expand(ms *matchState, s, p, ep []byte) []byte {
	//debug(fmt.Sprintf("min_expand(%q, %q, %q, %q)", ms, s, p, ep))
	for {
		res := match(ms, s, ep[1:])
		if res != nil {
			return res
		} else if len(s) > 0 && singlematch(s[0], p, ep) {
			// try with one more repetition
			s = s[1:]
		} else {
			return nil
		}
	}
	return nil
}

// Checks if a capture exists with the given capture index. Rather than
// providing an error, since we're not sure how we'd do that right now, we
// return -1 and handle error checking outside this routine.

func check_capture(ms *matchState, l int) int {
	//debug(fmt.Sprintf("check_capture(%q, %q)", ms, l))
	l = l - '1'
	if l < 0 || l >= ms.level || ms.capture[l].len == CAP_UNFINISHED {
		// error: invalid capture index
		return -1
	}
	return l
}

// Returns the first level that contains an unclosed capture, or -1 if there is
// no such capture level.

func capture_to_close(ms *matchState) int {
	//debug(fmt.Sprintf("capture_to_close(%q)", ms))
	for level := ms.level - 1; level >= 0; level = level - 1 {
		if ms.capture[level].len == CAP_UNFINISHED {
			return level
		}
	}
	panic("NO SOUP FOR YOU")
}

// Finds the end of a character class [] and return that part of the pattern

func classend(ms *matchState, p []byte) []byte {
	//debug(fmt.Sprintf("classend(%q, %q)", ms, p))
	var ch byte = p[0]
	p = p[1:]

	switch ch {
	case '%':
		{
			if len(p) == 0 {
				// error: malformed pattern, ends with '%'
				return nil
			}
			return p[1:]
		}
	case '[':
		{
			if p[0] == '^' {
				p = p[1:]
			}
			// look for a ']'
			for {
				if len(p) == 0 {
					// error: malformed pattern (missing ']')
					return nil
				}
				pch := p[0]
				p = p[1:]
				if pch == '%' && len(p) > 0 {
					// skip escapes (e.g. %])
					p = p[1:]
				}
				if p[0] == ']' {
					break
				}
			}
			return p[1:]
		}
	default:
		{
			return p
		}
	}

	return nil
}

// Sets up the match state to start a capture, and attempts to finish the match
// with that capture in place. If the further match fails, the capture is
// undone, otherwise the match is returned.

func start_capture(ms *matchState, s, p []byte, what int) []byte {
	//debug(fmt.Sprintf("start_capture(%q, %q, %q, %q)", ms, s, p, what))
	var res []byte
	var level int = ms.level

	if level >= LUA_MAXCAPTURES {
		// error: too many captures
		//debug("**** too many captures, getting out of here!")
		return nil
	}
	//debug(fmt.Sprintf("**** level: %d", level))
	ms.capture[level].src = s
	ms.capture[level].len = what
	ms.level = level + 1
	if res = match(ms, s, p); res == nil {
		//debug("undoing capture due to failed match")
		// match failed, so undo capture
		ms.level--
	}
	return res
}

// Ends a capture

func end_capture(ms *matchState, s, p []byte) []byte {
	//debug(fmt.Sprintf("end_capture(%q, %q, %q)", ms, s, p))
	var l int = capture_to_close(ms)
	if l == -1 {
		return nil
	}

	//debug(fmt.Sprintf("***\n\ns: %q, capsrc: %q\n\n", s, ms.capture[l].src))
	var res []byte
	// close the capture
	//debug(fmt.Sprintf("l: %d, capture: %q", l, ms.capture[l]))
	ms.capture[l].len = len(ms.capture[l].src) - len(s)
	if res = match(ms, s, p); res == nil {
		// undo the capture, remainder match failed
		ms.capture[l].len = CAP_UNFINISHED
	}
	return res
}

// Does the actual capture by checking the integrity of the capture state and
// copying the string into the capture slot. This function returns the
// remainder of the source string that was not captured.

func match_capture(ms *matchState, s []byte, l int) []byte {
	//debug(fmt.Sprintf("match_capture(%q, %q, %q)", ms, s, l))
	var clen int
	l = check_capture(ms, l)
	if l == -1 {
		return nil
	}

	clen = ms.capture[l].len
	//debug(fmt.Sprintf("clen: %d", clen))

	// ensure there is enough space in the source string to accommodate the
	// match
	if len(s)-clen >= 0 && bytes.Compare(ms.capture[l].src[0:clen], s[0:clen]) == 0 {
		return s[clen:]
	}

	return nil
}

// This function attempts to find a match for the pattern p in the string s
// starting at the first character. It does this by moving through the pattern
// changing the match state (and source/pattern strings) as necessary.

func match(ms *matchState, s, p []byte) []byte {
	//debug(fmt.Sprintf("match(%q, %q, %q)", ms, s, p))
init:
	//debug(fmt.Sprintf("match[init](%q, %q, %q)", ms, s, p))

	if len(p) == 0 {
		return s
	}

	var ep []byte
	var m bool

	switch p[0] {
	case '(':
		{ // start capture
			if p[1] == ')' { // position capture
				// TODO: We don't support these
				return nil
			} else {
				return start_capture(ms, s, p[1:], CAP_UNFINISHED)
			}
		}
	case ')':
		{ // end capture
			return end_capture(ms, s, p[1:])
		}
	case '%':
		{
			switch p[1] {
			case 'b':
				{ // balanced string
					s = matchbalance(ms, s, p[2:])
					if s == nil {
						return nil
					}
					p = p[4:]
				}
			// TODO: Support the frontier pattern
			default:
				{
					if isdigit(p[1]) { // capture result (%0-%9)
						s = match_capture(ms, s, int(p[1]))
						if s == nil {
							return nil
						}
						p = p[2:]
						goto init
					}
					goto dflt
				}
			}
		}
	case '$':
		{
			// check to ensure that the '$' is the last character in the pattern
			if len(p) == 1 {
				if len(s) == 0 {
					return s
				} else {
					return nil
				}
			} else {
				goto dflt
			}
		}
	default:
		goto dflt
	}

	goto skipdflt

dflt: // it is a pattern item
	ep = classend(ms, p) // points to what is next
	m = len(s) > 0 && singlematch(s[0], p, ep)
	//debug(fmt.Sprintf("m: %t", m))

	// Handle the case where ep has run out so we can't index it
	if len(ep) == 0 {
		if !m {
			return nil
		} else {
			s = s[1:]
			p = ep
			goto init
		}
	}

	switch ep[0] {
	case '?':
		{
			// If s has run out, the optional match passes
			if len(s) == 0 {
				return []byte{}
			}

			var res []byte = match(ms, s[1:], ep[1:])
			if m && res != nil {
				return res
			}
			p = ep[1:]
			goto init
		}
	case '*':
		{
			return max_expand(ms, s, p, ep)
		}
	case '+':
		{
			if m {
				return max_expand(ms, s[1:], p, ep)
			} else {
				return nil
			}
		}
	case '-':
		{
			return min_expand(ms, s, p, ep)
		}
	default:
		{
			if !m {
				return nil
			}
			s = s[1:]
			p = ep
			goto init
		}
	}

skipdflt:

	return nil
}

func get_onecapture(ms *matchState, i int, s, e []byte) []byte {
	//debug(fmt.Sprintf("get_onecapture(%q, %d, %q, %q)", ms, i, s, e))
	if i >= ms.level {
		if i == 0 {
			// return whole match
			return s[0 : len(s)-len(e)]
		} else {
			// error: invalid capture index
			return nil
		}
	} else {
		var l int = ms.capture[i].len
		if l == CAP_UNFINISHED {
			// error: unfinished capture
			return nil
		} else {
			return ms.capture[i].src[0:l]
		}
	}
	return nil
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
			init++ // 1st char is already checked by IndexBytes
			if bytes.Equal(s1[init-1:end], s2) {
				return init - 1
			} else { // find the next 'init' and try again
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

func add_s(ms *matchState, b *bytes.Buffer, s, e []byte, news []byte) {
	l := len(news)
	for i := 0; i < l; i++ {
		if news[i] != '%' {
			b.WriteByte(news[i])
		} else {
			i++ // skip ESC (%)
			if !isdigit(news[i]) {
				b.WriteByte(news[i])
			} else if news[i] == '0' {
				b.Write(s[0 : len(s)-len(e)])
			} else {
				cidx := int(news[i] - '1')
				b.Write(get_onecapture(ms, cidx, s, e))
			}
		}
	}
}

// Looks for the first match of pattern p in the string s. If it finds one,
// then match returns true and the captures from the pattern; otherwise it
// returns false, nil.  If pattern specifies no captures, then the whole match
// is returned.

func Match(s, p string) (bool, []string) {
	sb, pb := []byte(s), []byte(p)
	succ, _, _, caps := FindBytes(sb, pb, false)

	scaps := make([]string, len(caps))
	for idx, str := range caps {
		scaps[idx] = string(str)
	}

	return succ, scaps[0:len(caps)]
}

// Same as the Match function, however operates directly on byte arrays rather
// than strings. This package operates natively in bytes, so this function is
// called by Match to perform it's work.

func MatchBytes(s, p []byte) (bool, [][]byte) {
	succ, _, _, caps := FindBytes(s, p, false)
	return succ, caps
}

// Returns a channel that can be used to iterate over all the matches of
// pattern p in string s. The single value sent down this channel is an
// array of the captures from the match.

func Gmatch(s, p string) chan []string {
	out := make(chan []string)
	start := 0
	go func() {
		for {
			succ, _, e, caps := Find(s[start:], p, false)
			if !succ {
				close(out)
				return
			} else {
				out <- caps
				start = e + start
			}
		}
	}()

	return out
}

// Same as the Gmatch function, however operates directly on byte arrays rather
// than strings. This package operates natively in bytes, so this function is
// called by Gmatch to perform it's work.

func GmatchBytes(s, p []byte) chan [][]byte {
	out := make(chan [][]byte)
	start := 0
	go func() {
		for {
			succ, _, e, caps := FindBytes(s[start:], p, false)
			if succ {
				out <- caps
				start = e
			} else {
				close(out)
				return
			}
		}
	}()

	return out
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

	scaps := make([]string, len(caps))
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

	// Perform a find and capture, looping to potentially find a match later in
	// the string

	var anchor bool = false
	if p[0] == '^' {
		p = p[1:]
		anchor = true
	}

	ms := new(matchState)
	ms.capture = make([]capture, LUA_MAXCAPTURES, LUA_MAXCAPTURES)

	var init int = 0

	for {
		res := match(ms, s[init:], p)

		if res != nil {
			//debug(fmt.Sprintf("match res: %q", res))
			// Determine the start and end indices of the match
			var start int = init
			var end int = len(s) - len(res)

			// Fetch the captures
			captures := new([LUA_MAXCAPTURES][]byte)

			var i int
			var nlevels int
			if ms.level == 0 && len(s) > 0 {
				nlevels = 1
			} else {
				nlevels = ms.level
			}

			for i = 0; i < nlevels; i++ {
				captures[i] = get_onecapture(ms, i, s, res)
			}

			return true, start, end, captures[0:nlevels]
		} else if len(s)-init == 0 || anchor {
			break
		}

		init = init + 1
	}
	// No match found
	return false, -1, -1, nil
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
	//debug(fmt.Sprintf("ReplaceBytes(%q, %q, %q, %d)", src, patt, repl, max))
	var anchor bool = false

	if patt[0] == '^' {
		anchor = true
		patt = patt[1:]
	}

	var n int = 0
	var b bytes.Buffer
	ms := new(matchState)
	ms.src = src
	ms.capture = make([]capture, LUA_MAXCAPTURES, LUA_MAXCAPTURES)

	for n < max || max == -1 {
		//debug(fmt.Sprintf("loop: n: %d, src = %q, patt = %q, b: %q", n, src, patt, b.Bytes()))
		ms.level = 0
		e := match(ms, src, patt)
		//debug(fmt.Sprintf("** e: %q", e))
		if e != nil {
			n++
			//debug("Found a match, so replacing it")
			//debug(fmt.Sprintf("e: %q, b: %q", e, b.Bytes()))
			add_s(ms, &b, src, e, repl) // Use add_s directly here
			//debug(fmt.Sprintf("e: %q, b: %q", e, b.Bytes()))
		}
		//debug(fmt.Sprintf("src: %q, ms.src: %q", src, ms.src))
		if e != nil && len(src) > 0 { // Non empty match
			//debug("foo")
			src = e // skip it
		} else if len(src) > 0 {
			//debug("bar")
			b.WriteByte(src[0])
			src = src[1:]
		} else {
			//debug("baz")
			break
		}

		if anchor {
			break
		}
	}
	b.Write(src[0:])
	//debug(fmt.Sprintf("Replace complete: %q", b.Bytes()))
	return b.Bytes(), n
}
