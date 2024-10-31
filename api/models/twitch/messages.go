package twitch

import "time"

type BaseMessage struct {
	Metadata struct {
		MessageId        string    `json:"message_id"`
		MessageType      string    `json:"message_type"`
		MessageTimestamp time.Time `json:"message_timestamp"`
	} `json:"metadata"`
}

type WelcomeMessage struct {
	BaseMessage
	Payload struct {
		Session struct {
			Id                      string    `json:"id"`
			Status                  string    `json:"status"`
			ConnectedAt             time.Time `json:"connected_at"`
			KeepAliveTimeoutSeconds int       `json:"keep_alive_timeout_seconds"`
			ReconnectUrl            string    `json:"reconnect_url"`
		} `json:"session"`
	} `json:"payload"`
}

type NotificationMessage struct {
	BaseMessage
	Payload struct {
		Subscription struct {
			Id        string `json:"id"`
			Status    string `json:"status"`
			Type      string `json:"type"`
			Version   string `json:"version"`
			Cost      int    `json:"cost"`
			CreatedAt string `json:"created_at"`
			Condition struct {
				BroadcasterUserId string `json:"broadcaster_user_id"`
				UserId            string `json:"user_id"`
			} `json:"condition"`
			Transport any `json:"transport"`
		} `json:"subscription"`
		Event any `json:"event"`
	} `json:"payload"`
}

type ChannelPointsCustomRewardRedemptionAddEvent struct {
	Id                   string `json:"id"`
	BroadcasterUserId    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	UserId               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	UserInput            string `json:"user_input"`
	Status               string `json:"status"`
	Reward               struct {
		Id     string `json:"id"`
		Title  string `json:"title"`
		Cost   int    `json:"cost"`
		Prompt string `json:"prompt"`
	} `json:"reward"`
	RedeemedAt time.Time `json:"redeemed_at"`
}

type ChannelPollEndEvent struct {
	Id                   string `json:"id"`
	BroadcasterUserId    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	Title                string `json:"title"`
	Choices              []struct {
		Id                 string `json:"id"`
		Title              string `json:"title"`
		BitsVotes          int    `json:"bits_votes"`
		ChannelPointsVotes int    `json:"channel_points_votes"`
		Votes              int    `json:"votes"`
	} `json:"choices"`
	BitsVoting struct {
		IsEnabled     bool `json:"is_enabled"`
		AmountPerVote int  `json:"amount_per_vote"`
	} `json:"bits_voting"`
	ChannelPointsVoting struct {
		IsEnabled     bool `json:"is_enabled"`
		AmountPerVote int  `json:"amount_per_vote"`
	} `json:"channel_points_voting"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}
