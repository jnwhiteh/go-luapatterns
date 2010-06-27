package main

import "github.com/jnwhiteh/go-luapatterns"
import "strings"

func main() {
	teststr := strings.Repeat("a", 1e7) + "b"
	_, _ = luapatterns.Match(teststr, ".-b")
}
