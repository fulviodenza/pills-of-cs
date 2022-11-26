package bot

import (
	"encoding/json"
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

func parse() (SerializedPills, error) {

	f, err := os.Open(PILLS_ASSET)
	if err != nil {
		log.Fatal("[parse]: ", err)
		return SerializedPills{}, err
	}

	byteValue, err := io.ReadAll(f)
	if err != nil {
		return SerializedPills{}, err
	}

	sp := SerializedPills{}
	err = json.Unmarshal(byteValue, &sp)
	if err != nil {
		return SerializedPills{}, err
	}

	return sp, nil
}
