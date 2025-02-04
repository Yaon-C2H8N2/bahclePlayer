package twitch

import "time"

type UserInfoResponse struct {
	Data []UserInfo `json:"data"`
}

type UserInfo struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	Type            string `json:"type"`
	BroadcasterType string `json:"broadcaster_type"`
	Description     string `json:"description"`
	ProfileImageUrl string `json:"profile_image_url"`
	OfflineImageUrl string `json:"offline_image_url"`
	ViewCount       int    `json:"view_count"`
}

type SubscriptionResponse struct {
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

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type PollCreateResponse struct {
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

type RewardResponse struct {
	Data []Reward `json:"data"`
}

type Reward struct {
	BroadcasterName     string      `json:"broadcaster_name"`
	BroadcasterLogin    string      `json:"broadcaster_login"`
	BroadcasterId       string      `json:"broadcaster_id"`
	Id                  string      `json:"id"`
	Image               interface{} `json:"image"`
	BackgroundColor     string      `json:"background_color"`
	IsEnabled           bool        `json:"is_enabled"`
	Cost                int         `json:"cost"`
	Title               string      `json:"title"`
	Prompt              string      `json:"prompt"`
	IsUserInputRequired bool        `json:"is_user_input_required"`
	MaxPerStreamSetting struct {
		IsEnabled    bool `json:"is_enabled"`
		MaxPerStream int  `json:"max_per_stream"`
	} `json:"max_per_stream_setting"`
	MaxPerUserPerStreamSetting struct {
		IsEnabled           bool `json:"is_enabled"`
		MaxPerUserPerStream int  `json:"max_per_user_per_stream"`
	} `json:"max_per_user_per_stream_setting"`
	GlobalCooldownSetting struct {
		IsEnabled             bool `json:"is_enabled"`
		GlobalCooldownSeconds int  `json:"global_cooldown_seconds"`
	} `json:"global_cooldown_setting"`
	IsPaused     bool `json:"is_paused"`
	IsInStock    bool `json:"is_in_stock"`
	DefaultImage struct {
		Url1X string `json:"url_1x"`
		Url2X string `json:"url_2x"`
		Url4X string `json:"url_4x"`
	} `json:"default_image"`
	ShouldRedemptionsSkipRequestQueue bool        `json:"should_redemptions_skip_request_queue"`
	RedemptionsRedeemedCurrentStream  interface{} `json:"redemptions_redeemed_current_stream"`
	CooldownExpiresAt                 interface{} `json:"cooldown_expires_at"`
}
