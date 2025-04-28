package models

type OverlayType struct {
	OverlayTypeId int         `json:"overlay_type_id"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	Schema        interface{} `json:"schema"`
}
