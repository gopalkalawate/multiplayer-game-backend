package models

type Player struct{
	ID string `json:"id"`
	MMR int `json:"mmr"`
	Region string `json:"region"`
	Ping int `json:"ping"`
	JoinedAt int64 `json:"joined_at"`
}
