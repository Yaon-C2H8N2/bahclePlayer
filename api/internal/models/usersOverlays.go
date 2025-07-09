package models

import (
	"encoding/json"
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
)

type CommonUsersOverlays struct {
	OverlayTypeId int    `json:"overlay_type_id"`
	UserId        int    `json:"user_id"`
	Settings      string `json:"settings"`
}

func GetAllUsersOverlaysFromUserId(userId int) ([]CommonUsersOverlays, error) {
	conn := utils.GetConnection()
	defer conn.Release()

	sql := `
		SELECT overlay_type_id, user_id, settings
		FROM users_overlays
		WHERE user_id = $1
	`
	rows := utils.DoRequest(conn, sql, userId)
	var userOverlays []CommonUsersOverlays
	for rows.Next() {
		var userOverlay CommonUsersOverlays
		err := rows.Scan(&userOverlay.OverlayTypeId, &userOverlay.UserId, &userOverlay.Settings)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		userOverlays = append(userOverlays, userOverlay)
	}

	return userOverlays, nil
}

func GetUserOverlaySettingsByTwitchId(twitchId, overlayCode string) (string, error) {
	conn := utils.GetConnection()
	defer conn.Release()

	sql := `
		SELECT users_overlays.settings
		FROM users_overlays
		JOIN users ON users.user_id = users_overlays.user_id
		JOIN overlay_types ON overlay_types.overlay_type_id = users_overlays.overlay_type_id
		WHERE users.twitch_id = $1 AND overlay_types.overlay_code = $2
	`

	var settings string
	rows := utils.DoRequest(conn, sql, twitchId, overlayCode)
	if !rows.Next() {
		if rows.Err() != nil {
			return "", fmt.Errorf("failed to get user overlay settings: %w", rows.Err())
		}
		return "{}", nil
	}

	err := rows.Scan(&settings)
	if err != nil {
		return "", fmt.Errorf("failed to scan user overlay settings: %w", err)
	}

	return settings, nil
}

func SaveUserOverlaySettings(userId int, overlayCode string, settings interface{}) error {
	conn := utils.GetConnection()
	defer conn.Release()

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	sql := `
		INSERT INTO users_overlays (overlay_type_id, user_id, settings)
		VALUES ((SELECT overlay_type_id FROM overlay_types WHERE overlay_code = $1), $2, $3)
		ON CONFLICT (overlay_type_id, user_id)
		DO UPDATE SET settings = $3
		RETURNING overlay_type_id, user_id, settings
	`

	utils.DoRequest(conn, sql, overlayCode, userId, settingsJSON)
	return nil
}
