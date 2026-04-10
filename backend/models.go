package main

type transcribeResponse struct {
	Status string `json:"status"`
	Text   string `json:"text"`
}
