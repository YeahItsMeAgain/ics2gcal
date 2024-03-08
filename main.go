package main

import (
	"ics2gcal/config"
	"ics2gcal/gcal"
	"ics2gcal/ics"
	"ics2gcal/logger"
	"ics2gcal/updater"
	"log"
	"os"
	"time"
)

func main() {
	logger.Init()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "update":
			if err := updater.SelfUpdate(); err != nil {
				logger.Logger.Fatal(err)
			}
		}
		return
	}

	logger.Logger.Info("Initializing config")
	config.Init()

	logger.Logger.Info("Connecting to google calendar service")
	calendarSrv := gcal.Init()

	ticker := time.NewTicker(time.Duration(config.Config.UpdateIntervalMins) * time.Minute)
	for {
		logger.Logger.Info("Fetching events from webcal")
		events, err := ics.ParseFromWebcal(config.Config.WebcalURL)
		if err != nil {
			log.Fatalf("Failed to parse iCalendar data: %v", err)
		}
		
		logger.Logger.Info("Updating google calendar")
		gcal.PushIcalEventsToGcal(calendarSrv, events)

		logger.Logger.Info("Waiting for next tick")
		<- ticker.C
	}
}
