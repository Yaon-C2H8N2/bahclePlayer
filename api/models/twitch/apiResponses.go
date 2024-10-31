package twitch

import "time"

type userInfoResponse struct {
	Data []struct {
		ID              string `json:"id"`
		Login           string `json:"login"`
		DisplayName     string `json:"display_name"`
		Type            string `json:"type"`
		BroadcasterType string `json:"broadcaster_type"`
		Description     string `json:"description"`
		ProfileImageUrl string `json:"profile_image_url"`
		OfflineImageUrl string `json:"offline_image_url"`
		ViewCount       int    `json:"view_count"`
	} `json:"data"`
}

type subcriptionResponse struct {
	Data []struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		Type      string `json:"type"`
		Version   string `json:"version"`
		Condition any    `json:"condition"`
		Transport struct {
			Method     string `json:"method"`
			Callback   string `json:"callback"`
			Secret     string `json:"secret"`
			SesssionId string `json:"session_id"`
		} `json:"transport"`
		CreatedAt string `json:"created_at"`
	} `json:"data"`
	Total        int `json:"total"`
	TotalCost    int `json:"total_cost"`
	MaxTotalCost int `json:"max_total_cost"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type pollCreateResponse struct {
	Data []struct {
		Id               string `json:"id"`
		BroadcasterId    string `json:"broadcaster_id"`
		BroadcasterName  string `json:"broadcaster_name"`
		BroadcasterLogin string `json:"broadcaster_login"`
		Title            string `json:"title"`
		Choices          []struct {
			Id                 string `json:"id"`
			Title              string `json:"title"`
			Votes              int    `json:"votes"`
			ChannelPointsVotes int    `json:"channel_points_votes"`
			BitsVotes          int    `json:"bits_votes"`
		} `json:"choices"`
		BitsVotingEnabled          bool      `json:"bits_voting_enabled"`
		BitsPerVote                int       `json:"bits_per_vote"`
		ChannelPointsVotingEnabled bool      `json:"channel_points_voting_enabled"`
		ChannelPointsPerVote       int       `json:"channel_points_per_vote"`
		Status                     string    `json:"status"`
		Duration                   int       `json:"duration"`
		StartedAt                  time.Time `json:"started_at"`
	} `json:"data"`
}
