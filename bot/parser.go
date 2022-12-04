package bot

import (
	"io"
	"log"
	"os"
)

const PILLS_ASSET = "./assets/pills.json"

type SerializedPills struct {
	Pills []Pill `json:"pills"`
}

type Pill struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func parse(dst *[]byte) ([]byte, error) {

	f, err := os.Open(PILLS_ASSET)
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
