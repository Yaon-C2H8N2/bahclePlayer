package twitch

type Condition struct {
	BroadcasterUserId string `json:"broadcaster_user_id"`
	UserId            string `json:"user_id"`
}

type Transport struct {
	Method    string `json:"method"`
	SessionId string `json:"session_id"`
}

type SubscriptionRequest struct {
	Type      string    `json:"type"`
	Version   string    `json:"version"`
	Condition Condition `json:"condition"`
	Transport Transport `json:"transport"`
}

type AppTokenRequest struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
}

type PollCreateRequest struct {
	BroadcasterId string `json:"broadcaster_id"`
	Title         string `json:"title"`
	Choices       []struct {
		Title string `json:"title"`
	} `json:"choices"`
	ChannelPointsVotingEnabled bool `json:"channel_points_voting_enabled"`
	ChannelPointsPerVote       int  `json:"channel_points_per_vote"`
	Duration                   int  `json:"duration"`
}
