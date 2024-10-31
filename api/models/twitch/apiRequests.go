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

type pollCreateRequest struct {
	BroadcasterId string `json:"broadcaster_id"`
	Title         string `json:"title"`
	Choices       []struct {
		Title string `json:"title"`
	} `json:"choices"`
	ChannelPointsVotingEnabled bool `json:"channel_points_voting_enabled"`
	ChannelPointsPerVote       int  `json:"channel_points_per_vote"`
	Duration                   int  `json:"duration"`
}
