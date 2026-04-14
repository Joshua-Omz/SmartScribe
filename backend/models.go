package main

type GoogleSTTResponse struct {
    Results []GoogleSTTResult `json:"results"`
}

type GoogleSTTResult struct {
    Alternatives []GoogleSTTAlternative `json:"alternatives"`
}

type GoogleSTTAlternative struct {
    Transcript string  `json:"transcript"`
    Confidence float64 `json:"confidence"`
}
type TranscriptionResponse struct {
	Status string `json:"status"`
	Text   string `json:"text"`
}

type TranscriptionResponse2 struct {
	Status     string `json:"status"`
	RawText    string `json:"raw_text"`
	Structured SOAP   `json:"structured_data"`
	ErrorMsg   string `json:"error,omitempty"` // Omitted if empty
}

type SOAP struct {
	Subjective string `json:"subjective"`
	Objective  string `json:"objective"`
	Assessment string `json:"assessment"`
	Plan       string `json:"plan"`
}
