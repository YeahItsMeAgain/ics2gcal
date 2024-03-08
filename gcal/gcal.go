package gcal

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"ics2gcal/config"
	"ics2gcal/logger"
	"log"
	"time"

	ical "github.com/arran4/golang-ical"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

func Init() *calendar.Service{
	ctx := context.Background()
	srv, err := calendar.NewService(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		log.Fatalf("Unable to retrieve calendar client: %v", err)
	}

	// For some reason service accounts need to "Accept" calendar invitations.
	entry := calendar.CalendarListEntry{Id: config.Config.CalendarId}
	srv.CalendarList.Insert(&entry).Do()
	return srv

}
func icalIdToGcalId(icalId string) string {
	idMd5 := md5.Sum([]byte(icalId))
	return hex.EncodeToString(idMd5[:])
}

func IcalEventsToGcal(calendarSrv *calendar.Service, events []*ical.VEvent) {
	for _, event := range events {
		eventSummary := event.GetProperty(ical.ComponentPropertySummary).Value
		startTime, startTimeErr := event.GetStartAt()
		endTime, endTimeErr := event.GetEndAt()
		if startTimeErr != nil || endTimeErr != nil {
			continue
		}

		// TODO: reminders
		// TODO: periodic updates
		gevent := &calendar.Event{
			Summary: eventSummary,
			Start: &calendar.EventDateTime{
				DateTime: startTime.Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: endTime.Format(time.RFC3339),
			},
			Description: event.GetProperty(ical.ComponentPropertyDescription).Value,
			Id:          icalIdToGcalId(event.GetProperty(ical.ComponentPropertyUniqueId).Value),
		}

		_, err := calendarSrv.Events.Insert(config.Config.CalendarId, gevent).Do()
		if err != nil {
			if err.(*googleapi.Error).Code == 409 {    // Already Exists.
				_, err := calendarSrv.Events.Update(config.Config.CalendarId, gevent.Id, gevent).Do()
				if err != nil {
					logger.Logger.Errorf("Unable to update event: %s. %v\n", eventSummary, err)
					continue
				}
			} else {
				logger.Logger.Errorf("Unable to create event: %s. %v\n", eventSummary, err)
				continue
			}
		}
	}
	logger.Logger.Info("Finished updating events")
}
