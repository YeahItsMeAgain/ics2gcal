package main

import (
	"ics2gcal/config"
	"ics2gcal/gcal"
	"ics2gcal/ics"
	"ics2gcal/logger"
	"log"
)

func main() {
	logger.Init()

	logger.Logger.Info("Initializing config")
	config.Init()

	logger.Logger.Info("Connecting to google calendar service")
	calendarSrv := gcal.Init()

	logger.Logger.Info("Fetching events from webcal")
	events, err := ics.ParseFromWebcal(config.Config.WebcalURL)
	if err != nil {
		log.Fatalf("Failed to parse iCalendar data: %v", err)
	}

	logger.Logger.Info("Updating google calendar")
	gcal.IcalEventsToGcal(calendarSrv, events)
}
