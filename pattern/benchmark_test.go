package luapatterns

import "testing"
import "strings"

var limit int = 1e6
var longBytes = make([]byte, limit + 1, limit + 1)
var longString = strings.Repeat("a", limit) + "b"

func BenchmarkLongBytes(b *testing.B) {
	for i := 0; i < limit; i++ {
		longBytes[i] = 'a'
	}
	longBytes[limit] = 'b'
	_, _, _, _ = FindBytes(longBytes, []byte(".-b"), false)
}

func BenchmarkLongString(b *testing.B) {
	_, _, _, _ = Find(longString, ".-b", false)
}
