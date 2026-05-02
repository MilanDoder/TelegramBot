package main

type KvizPitanje struct {
	Pitanje  string
	Opcije   []string
	Tacan    int
}

type KvizStanje struct {
	Pitanja  []KvizPitanje
	Trenutno int
	Poeni    int
	Aktivno  bool
}

// OpenTrivia API response strukture
type OTDBResponse struct {
	ResponseCode int          `json:"response_code"`
	Results      []OTDBResult `json:"results"`
}

type OTDBResult struct {
	Question         string   `json:"question"`
	CorrectAnswer    string   `json:"correct_answer"`
	IncorrectAnswers []string `json:"incorrect_answers"`
}

var kvizovi = map[int64]*KvizStanje{}