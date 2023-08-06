package utils

import (
	"math/rand"
	"strconv"
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

func AggregateTags(tags []string) string {
	msg := ""
	for _, s := range tags {
		msg += "- " + s + "\n"
	}

	return msg
}

func ValidateTime(times []string, tz string) bool {
	// the only accepted format is HH:MM, so, with 2 elements in the times array
	if len(times) != 2 {
		return false
	}

	hours, err := strconv.Atoi(times[0])
	if err != nil {
		return false
	}
	minutes, err := strconv.Atoi(times[1])
	if err != nil {
		return false
	}

	// validate timezone
	_, err = time.LoadLocation(tz)
	if err != nil {
		return false
	}

	if hours < 0 || hours >= 24 {
		return false
	}
	if minutes < 0 || minutes >= 60 {
		return false
	}

	return true
}
