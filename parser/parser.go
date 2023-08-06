package parser

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/pills-of-cs/utils"
)

func Parse(filename string, dst *[]byte) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal("[parse]: ", err)
		return []byte{}, err
	}

	bytes, err := io.ReadAll(f)
	if err != nil {
		return []byte{}, err
	}

	*dst = bytes

	return *dst, nil
}

func ParseSchedule(s string) (string, error) {
	// times contains an array with two elements [Hours, Minutes]
	times := strings.SplitN(s, ":", -1)
	// in the crontab minutes come as first field
	if !utils.ValidateTime(times) {
		return "", fmt.Errorf("error validating time")
	}
	return fmt.Sprintf("%s %s * * *", times[1], times[0]), nil
}
