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
	calendarSrv := gcal.GetCalendarSrv()

	events, err := ics.ParseFromWebcal(config.Config.WebcalURL)
	if err != nil {
		log.Fatalf("Failed to parse iCalendar data: %v", err)
	}

	gcal.IcalEventsToGcal(calendarSrv, events)
}
