package twitch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	request := &appTokenRequest{
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

	var tokenResponse = &tokenResponse{}
	err = json.Unmarshal(body, tokenResponse)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

func (aw *ApiWrapper) GetBroadcasterIdFromToken(userToken string) (string, error) {
	twitchUrl := "https://api.twitch.tv/helix/users"

	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", twitchUrl, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+userToken)
	req.Header.Add("Client-Id", aw.clientId)

	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var userInfoResponse = &userInfoResponse{}
	err = json.Unmarshal(body, userInfoResponse)
	if err != nil {
		return "", err
	}

	return userInfoResponse.Data[0].ID, nil
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

	data := pollCreateRequest{
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

	pollResponse := &pollCreateResponse{}
	err = json.Unmarshal(body, pollResponse)
	if err != nil {
		return "", err
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
