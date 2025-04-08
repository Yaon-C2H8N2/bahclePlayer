package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models/youtube"
	"io"
	"net/http"
	"os"
)

func GetVideoInfo(id string) (youtube.VideoInfoResponse, error) {
	youtubeUrl := "https://www.googleapis.com/youtube/v3/videos"
	youtubeUrl += "?id=" + id + "&part=snippet&part=contentDetails" + "&key=" + os.Getenv("YOUTUBE_API_KEY")

	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", youtubeUrl, nil)
	if err != nil {
		return youtube.VideoInfoResponse{}, err
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return youtube.VideoInfoResponse{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return youtube.VideoInfoResponse{}, err
	}

	var videoInfoResponse = youtube.VideoInfoResponse{}
	err = json.Unmarshal(body, &videoInfoResponse)
	if err != nil {
		return youtube.VideoInfoResponse{}, err
	}
	if videoInfoResponse.Items == nil {
		return youtube.VideoInfoResponse{}, fmt.Errorf("Failed to get video info")
	}
	return videoInfoResponse, nil
}
