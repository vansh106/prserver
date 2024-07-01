package main

import (
	"context"
	"fmt"
	// "log"
	"net/http"
	"os"
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

	r.Run("0.0.0.0:" + os.Getenv("PORT"))
}

func getRooms(c *gin.Context) {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("./creds2.json"))
	if err != nil {
		fmt.Println("Unable to retrieve Sheets client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error   sheet": err})

	}
	fmt.Println("here")

	spreadsheetId := "1eivXSkSQs37JfPOQ6YME2piD71XxbW73mq5fdKhkwM4"
	readRange := "Rooms!A2:K"

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error retrieving from sheet": err})
		return
	}

	if len(resp.Values) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No data found."})
		return
	}

	// Get query parameters
	availability := c.DefaultQuery("Availability", "Available")

	var filteredRooms []Room
	for _, row := range resp.Values {
		// if len(row) < 12 {
		// 	continue // Skip rows with insufficient data
		// }

		fmt.Println(row)
		if matchesFilters(row, c) && row[9] == availability {

			imageUrls := []string{}
			if row[10] != nil {
				imageUrls = strings.Split(row[10].(string), ",")
			}
			fmt.Println("[]", imageUrls)
			filteredRooms = append(filteredRooms, Room{
				ImageUrls: imageUrls,
				Area:      row[5].(string),
				Address:   row[4].(string),
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
		"Availablity":      9,
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
