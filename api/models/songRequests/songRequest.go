package songRequests

type SongRequest struct {
	TwitchPollID       string `json:"twitch_poll_id"`
	TwitchRedemptionID string `json:"twitch_redemption_id"`
	TwitchRewardID     string `json:"twitch_reward_id"`
	YoutubeID          string `json:"youtube_id"`
	Title              string `json:"title"`
	Channel            string `json:"channel"`
	Duration           int    `json:"duration"`
	Thumbnail          string `json:"thumbnail"`
}
