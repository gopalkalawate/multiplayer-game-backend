package models

type Match struct {
	ID      string   `json:"id"`
	Players []string `json:"players"`
	Region  string   `json:"region"`
	Status  string   `json:"status"`
}
