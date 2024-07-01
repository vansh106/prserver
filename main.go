package main

import (
	"context"
	// "fmt"
	// "log"
	"net/http"
	// "os"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Room struct {
	ImageUrls []string `json:"imageUrls"`
	Area      string   `json:"area"`
	Address   string   `json:"address"`
}

func main() {
	r := gin.Default()

	r.GET("/rooms", getRooms)

	r.Run(":8080")
}

func getRooms(c *gin.Context) {
    ctx := context.Background()
    srv, err := sheets.NewService(ctx, option.WithCredentialsFile("./creds2.json"))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve Sheets client: " + err.Error()})
        return
    }

    spreadsheetId := "1u6NPfLncyDlqeSfFq-DB3csjSQTaFTAUePrU0eNXlJY"
    readRange := "Rooms!A2:K"

    resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving from sheet: " + err.Error()})
        return
    }

    if len(resp.Values) == 0 {
        c.JSON(http.StatusNotFound, gin.H{"message": "No data found."})
        return
    }

    availability := c.DefaultQuery("Availability", "Available")

    var filteredRooms []Room
    for _, row := range resp.Values {
        if len(row) < 11 {
            continue // Skip rows with insufficient data
        }
        if matchesFilters(row, c) && row[9] == availability {
            imageUrls := []string{}
            if imageUrlStr, ok := row[10].(string); ok {
                imageUrls = strings.Split(imageUrlStr, ",")
            }
            area, _ := row[5].(string)
            address, _ := row[4].(string)
            filteredRooms = append(filteredRooms, Room{
                ImageUrls: imageUrls,
                Area:      area,
                Address:   address,
            })
        }
    }

    c.JSON(http.StatusOK, filteredRooms)
}
func matchesFilters(row []interface{}, c *gin.Context) bool {

	filters := map[string]int{
		"Type":             2,
		"Occupancy":        3,
		"Area":             5,
		"BachelorsAllowed": 6,
		"Furnished":        7,
		"Gender":           8,
		"Availability":      9,
	}

	for param, index := range filters {
		if value := c.Query(param); value != "" {
			if param == "Allowed" {
				if !strings.Contains(row[index].(string), value) {
					return false
				}
			} else if row[index].(string) != value {
				return false
			}
		}
	}

	return true
}
