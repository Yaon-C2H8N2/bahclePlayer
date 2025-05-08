package models

import (
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
)

type OverlayType struct {
	OverlayTypeId int         `json:"overlay_type_id"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	Schema        interface{} `json:"schema"`
	OverlayCode   string      `json:"overlay_code"`
}

func GetAllOverlayTypes() ([]OverlayType, error) {
	conn := utils.GetConnection()
	defer conn.Release()

	sql := `
		SELECT overlay_type_id, name, description, schema, overlay_code
		FROM overlay_types
	`
	rows := utils.DoRequest(conn, sql)
	var overlayTypes []OverlayType
	for rows.Next() {
		var overlayType OverlayType
		err := rows.Scan(&overlayType.OverlayTypeId, &overlayType.Name, &overlayType.Description, &overlayType.Schema, &overlayType.OverlayCode)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		overlayTypes = append(overlayTypes, overlayType)
	}

	return overlayTypes, nil
}
