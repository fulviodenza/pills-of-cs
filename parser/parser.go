package parser

import (
	"io"
	"log"
	"os"
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
