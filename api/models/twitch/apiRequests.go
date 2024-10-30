package twitch

type condition struct {
	BroadcasterUserId string `json:"broadcaster_user_id"`
	UserId            string `json:"user_id"`
}

type transport struct {
	Method    string `json:"method"`
	SessionId string `json:"session_id"`
}

type subscriptionRequest struct {
	Type      string    `json:"type"`
	Version   string    `json:"version"`
	Condition condition `json:"condition"`
	Transport transport `json:"transport"`
}

type appTokenRequest struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
}
