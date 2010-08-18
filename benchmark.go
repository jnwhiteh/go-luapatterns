package main

import "luapatterns"
import "strings"

func main() {
	teststr := strings.Repeat("a", 1e8) + "b"
	_, _ = luapatterns.Match(teststr, ".-b")
}
