package luapatterns

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

type matchState struct {
}

func Find(s, p []byte) (bool, int, int, []byte) {
	var anchor bool = false
	if p[0] == '^' {
		p = p[1:]
		anchor = true
	}

	ms := new(matchState)

	for {
		res := match(ms, s, p)
		if res != nil {
			// TODO: Fetch and return captures
			// Match was successful
			return true, 0, 0, nil
		} else if len(s) == 0 || anchor {
			break
		}
	}
	// No match found
	return false, -1, -1, nil
}

func tolower(c byte) byte {
	if c >= 65 && c <= 90 {				// upper case
		return c + 32
	}
	return c
}

// Returns whether or not a byte matches a specified character class, which may
// simply be a non-special character itself.

func match_class(c byte, cl byte) (res bool) {
	cllower := tolower(cl)
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

	return !res		// handle upper-case reverse classes
}

// Returns whether or not a single character matches the pattern currently
// being examined. It needs to 'lookahead' in order to accomplish this, for
// example when the pattern is a character class like %l. This is the purpose
// of the argument ep.

func singlematch(c byte, p []byte) bool {
	switch p[0] {
		case '.': return true
		case '%': {
			if len(p) < 2 {
				return false
			}
			return match_class(c, p[1])
		}
		default: return p[0] == c
	}

	return false
}

// This function attempts to find a match for the pattern p in the string s
// starting at the first character. It does this by moving through the pattern
// changing the match state (and source/pattern strings) as necessary.
func match(ms *matchState, s, p []byte) []byte {
	for {
		switch p[0] {
			case '%':
			case '$':
			default: {
			}
		}
	}

	return nil
}
