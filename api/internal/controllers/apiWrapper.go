package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models/twitch"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type ApiWrapper struct {
	appToken string
	clientId string
}

func GetApiWrapper() *ApiWrapper {
	return &ApiWrapper{}
}

func RequestAppToken(clientId string, clientSecret string) (string, error) {
	twitchUrl := "https://id.twitch.tv/oauth2/token"

	request := &twitch.AppTokenRequest{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		GrantType:    "client_credentials",
	}
	twitchUrl += "?client_id=" + request.ClientId + "&client_secret=" + request.ClientSecret + "&grant_type=" + request.GrantType

	httpClient := &http.Client{}
	req, err := http.NewRequest("POST", twitchUrl, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return "", err
	}

	var tokenResponse = &twitch.TokenResponse{}
	err = json.Unmarshal(body, tokenResponse)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

func cacheToken(token *twitch.UserTokenResponse) error {
	key := "auth:token:" + token.AccessToken
	ctx := context.Background()
	rdb := utils.GetValkeyClient()

	err := rdb.HSet(ctx, key, map[string]interface{}{
		"token":         token.AccessToken,
		"refresh_token": token.RefreshToken,
		"expiry":        token.ExpiresIn,
	}).Err()

	if err != nil {
		return err
	}

	return rdb.Expire(ctx, key, 3*time.Hour).Err()
}

func RequestUserToken(code string) (*twitch.UserTokenResponse, error) {
	twitchUrl := "https://id.twitch.tv/oauth2/token"
	appUrl := os.Getenv("APP_URL")
	clientId := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")

	request := &twitch.TokenFromCodeRequest{
		TokenRequest: twitch.TokenRequest{
			ClientId:     clientId,
			ClientSecret: clientSecret,
			GrantType:    "authorization_code",
		},
		Code:        code,
		RedirectUri: appUrl,
	}
	requestBody := strings.NewReader("client_id=" + request.ClientId + "&client_secret=" + request.ClientSecret + "&code=" + request.Code + "&grant_type=" + request.GrantType + "&redirect_uri=" + request.RedirectUri)

	httpClient := &http.Client{}
	req, err := http.NewRequest("POST", twitchUrl, requestBody)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	var tokenResponse = &twitch.UserTokenResponse{}
	err = json.Unmarshal(body, tokenResponse)
	if err != nil {
		return nil, err
	}

	err = cacheToken(tokenResponse)
	if err != nil {
		return nil, err
	}

	return tokenResponse, nil
}

func RefreshUserToken(refreshToken string) (*twitch.UserTokenResponse, error) {
	twitchUrl := "https://id.twitch.tv/oauth2/token"
	clientId := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")

	request := &twitch.TokenFromRefreshRequest{
		TokenRequest: twitch.TokenRequest{
			ClientId:     clientId,
			ClientSecret: clientSecret,
			GrantType:    "refresh_token",
		},
		RefreshToken: refreshToken,
	}
	requestBody := strings.NewReader("client_id=" + request.ClientId + "&client_secret=" + request.ClientSecret + "&grant_type=" + request.GrantType + "&refresh_token=" + request.RefreshToken)

	httpClient := &http.Client{}
	req, err := http.NewRequest("POST", twitchUrl, requestBody)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	var tokenResponse = &twitch.UserTokenResponse{}
	err = json.Unmarshal(body, tokenResponse)
	if err != nil {
		return nil, err
	}

	err = cacheToken(tokenResponse)
	if err != nil {
		return nil, err
	}

	return tokenResponse, nil
}

func (aw *ApiWrapper) GetUserInfoFromToken(userToken string) (twitch.UserInfo, error) {
	twitchUrl := "https://api.twitch.tv/helix/users"

	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", twitchUrl, nil)
	if err != nil {
		return twitch.UserInfo{}, err
	}

	req.Header.Add("Authorization", "Bearer "+userToken)
	req.Header.Add("Client-Id", aw.clientId)

	res, err := httpClient.Do(req)
	if err != nil {
		return twitch.UserInfo{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return twitch.UserInfo{}, err
	}
	var userInfoResponse = &twitch.UserInfoResponse{}
	err = json.Unmarshal(body, userInfoResponse)
	if err != nil {
		return twitch.UserInfo{}, err
	}

	if len(userInfoResponse.Data) == 0 {
		return twitch.UserInfo{}, fmt.Errorf("failed to get user info : %s", string(body))
	}

	return userInfoResponse.Data[0], nil
}

func (aw *ApiWrapper) SetAppToken(token string) {
	aw.appToken = token
}

func (aw *ApiWrapper) GetAppToken() string {
	return aw.appToken
}

func (aw *ApiWrapper) SetClientId(clientId string) {
	aw.clientId = clientId
}

func (aw *ApiWrapper) CreatePoll(userToken string, broadcasterId string, title string, choices []string, duration int) (string, error) {
	twitchUrl := "https://api.twitch.tv/helix/polls"

	var choicesData []struct {
		Title string `json:"title"`
	}
	for _, choice := range choices {
		var choiceData struct {
			Title string `json:"title"`
		}
		choiceData.Title = choice
		choicesData = append(choicesData, choiceData)
	}

	data := twitch.PollCreateRequest{
		BroadcasterId:              broadcasterId,
		Title:                      title,
		Choices:                    choicesData,
		ChannelPointsVotingEnabled: false,
		ChannelPointsPerVote:       0,
		Duration:                   duration,
	}

	httpClient := &http.Client{}
	bytes, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", twitchUrl, strings.NewReader(string(bytes)))
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+userToken)
	req.Header.Add("Client-Id", aw.clientId)
	req.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	pollResponse := &twitch.PollCreateResponse{}
	err = json.Unmarshal(body, pollResponse)
	if err != nil {
		return "", err
	}

	if len(pollResponse.Data) == 0 {
		return "", fmt.Errorf("failed to create poll : %s", string(body))
	}

	return pollResponse.Data[0].Id, nil
}

func (aw *ApiWrapper) UpdateRedemptionStatus(userToken string, redemptionId string, broadcasterId string, rewardId string, newStatus string) error {
	twitchUrl := fmt.Sprintf("https://api.twitch.tv/helix/channel_points/custom_rewards/redemptions?id=%s&broadcaster_id=%s&reward_id=%s", redemptionId, broadcasterId, rewardId)

	body := strings.NewReader(`{"status":"` + newStatus + `"}`)

	httpClient := &http.Client{}
	req, err := http.NewRequest("PATCH", twitchUrl, body)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+userToken)
	req.Header.Add("Client-Id", aw.clientId)
	req.Header.Add("Content-Type", "application/json")

	_, err = httpClient.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func (aw *ApiWrapper) GetChannelRewards(userToken string, broadcasterId string) ([]twitch.Reward, error) {
	twitchUrl := fmt.Sprintf("https://api.twitch.tv/helix/channel_points/custom_rewards?broadcaster_id=%s", broadcasterId)

	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", twitchUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+userToken)
	req.Header.Add("Client-Id", aw.clientId)

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	rewardsResponse := &twitch.RewardResponse{}
	err = json.Unmarshal(body, rewardsResponse)
	if err != nil {
		return nil, err
	}

	return rewardsResponse.Data, nil
}
