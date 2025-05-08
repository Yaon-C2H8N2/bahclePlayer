package models

import (
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
		WHERE users.id = $1
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
