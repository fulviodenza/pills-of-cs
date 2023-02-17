package utils

import (
	"math/rand"
	"time"
)

func MakeTimestamp(len int) int64 {
	millisec := int64(time.Millisecond)
	now := time.Now().UnixNano()
	division := now / millisec
	return (division) % int64(len)
}

func Pick[K comparable, V any](m map[K]V) V {
	k := rand.Intn(len(m))
	for _, x := range m {
		if k == 0 {
			return x
		}
		k--
	}
	panic("unreachable")
}
