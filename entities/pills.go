package entities

type SerializedPills struct {
	Pills []Pill `json:"pills"`
}

type Pill struct {
	Title string   `json:"title"`
	Body  string   `json:"body"`
	Tags  []string `json:"tags"`
}
