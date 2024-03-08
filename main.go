package main

import (
	"fmt"
	"ics2gcal/config"
	"ics2gcal/ics"
	"ics2gcal/logger"
	"log"

	ical "github.com/arran4/golang-ical"
)

func main() {
	logger.Init()

	logger.Logger.Info("Initializing config")
	config.Init()

	// Fetching the iCalendar file
	resp, err := ics.FetchWebcal(config.Config.WebcalURL)
	if err != nil {
		log.Fatalf("Failed to fetch iCalendar file: %v", err)
	}
	defer resp.Body.Close()

	// Parsing the iCalendar data
	events, err := ics.ParseICS(resp.Body)
	if err != nil {
		log.Fatalf("Failed to parse iCalendar data: %v", err)
	}

	// Printing parsed events
	for _, event := range events {
		fmt.Printf("Event: %s\n", event.GetProperty(ical.ComponentPropertySummary).Value)
		fmt.Println("----------")
	}
}
