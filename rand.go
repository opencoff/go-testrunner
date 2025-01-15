// rand.go - handy random bytes/ints collection

package testrunner

import (
	"crypto/rand"
	"fmt"
	"golang.org/x/exp/constraints"
)

// fill buffer 'buf' with random bytes
func randBytes(buf []byte) {
	_, err := rand.Read(buf)
	if err != nil {
		panicf("rand: can't read %d bytes: %s", len(buf), err)
	}
}

// make a new buffer of 'n' bytes and fill it with
// random bytes
func randBuf[T constraints.Integer](n T) []byte {
	b := make([]byte, n)
	randBytes(b)
	return b
}

func randstr(n int) string {
	b := randBuf(n)
	return fmt.Sprintf("%x", b)
}
