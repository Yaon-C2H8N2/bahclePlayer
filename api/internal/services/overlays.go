package services

import (
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
	"github.com/gin-gonic/gin"
)

func getOverlays(c *gin.Context) {
	conn := utils.GetConnection()
	defer conn.Release()

	sql := `
		SELECT overlay_type_id, name, description, schema
		FROM overlay_types
	`
	rows := utils.DoRequest(conn, sql)
	var overlayTypes []models.OverlayType
	for rows.Next() {
		var overlayType models.OverlayType
		err := rows.Scan(&overlayType.OverlayTypeId, &overlayType.Name, &overlayType.Description, &overlayType.Schema)
		if err != nil {
			fmt.Println("Failed to get overlay types:", err)
			c.JSON(500, gin.H{
				"error": "Failed to get overlay types",
			})
			return
		}
		overlayTypes = append(overlayTypes, overlayType)
	}

	c.JSON(200, gin.H{
		"overlay_types": overlayTypes,
	})
}
