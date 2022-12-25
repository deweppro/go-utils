package random

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	digest = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func Bytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = digest[rand.Intn(len(digest))]
	}
	return b
}
