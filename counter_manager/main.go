package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()
	router.GET("/get_dump", func(ctx *gin.Context) {

		dirEntries, err := os.ReadDir("../output_dumps")
		if err != nil {
			log.Println("unable to access directory: ../output_dumps")
			return
		}
		timeNow := time.Now()
		parsedDatetimes := []time.Time{}

		// find the entry with the minimum difference to the current time
		for _, entry := range dirEntries {
			if !strings.Contains(entry.Name(), ".json") {
				continue
			}
			// take off .json suffix from the filename
			entryNameNoSuffix := entry.Name()[0 : len(entry.Name())-5]

			parsedDatetimeItem, err := time.Parse("2006-01-02T15:04:05Z", entryNameNoSuffix)
			if err != nil {
				log.Printf("unable to parse filename: %s as datetime string, the suffix removed string was: %s",
					entry.Name(),
					entryNameNoSuffix,
				)
			}
			parsedDatetimes = append(parsedDatetimes, parsedDatetimeItem)
		}

		closestDatetime := time.Time{}
		// setting the starting value for the duration to 10 years, so that all reasonable values are less than that
		shortestDiff := time.Duration(1) * time.Hour * 24 * 365 * 10

		for _, datetimeItem := range parsedDatetimes {

			// only consider data dumped prior to the start of the API call
			if datetimeItem.After(timeNow) {
				continue
			}
			if timeNow.Sub(datetimeItem) < shortestDiff {
				shortestDiff = timeNow.Sub(datetimeItem)
				closestDatetime = datetimeItem
			}
		}

		filepath := fmt.Sprintf("../output_dumps/%s.json", closestDatetime.Format("2006-01-02T15:04:05Z"))
		log.Printf("found the most recent dump json file: %s", filepath)
		// output file as response
		ctx.File(filepath)
	})

	router.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
