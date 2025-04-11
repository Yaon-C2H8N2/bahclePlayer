package models

type AppStatus struct {
	TwitchClientId string `json:"twitch_client_id"`
	AppUrl         string `json:"app_url"`
	Started        bool   `json:"started"`
	Version        string `json:"version"`
}
