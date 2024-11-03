package models

import "time"

type UsersVideos struct {
	VideoId      int       `json:"video_id"`
	UserId       int       `json:"user_id"`
	YoutubeId    string    `json:"youtube_id"`
	Url          string    `json:"url"`
	Title        string    `json:"title"`
	Duration     string    `json:"duration"`
	Type         string    `json:"type"`
	CreatedAt    time.Time `json:"created_at"`
	ThumbnailUrl string    `json:"thumbnail_url"`
	AddedBy      string    `json:"added_by"`
}
