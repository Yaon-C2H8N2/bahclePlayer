package models

type VideoAddRequest struct {
	Url  string `json:"url"`
	Type string `json:"type"`
}
