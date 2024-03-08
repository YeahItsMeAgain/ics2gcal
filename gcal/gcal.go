package gcal

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"ics2gcal/config"
	"ics2gcal/logger"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	ical "github.com/arran4/golang-ical"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	logger.Logger.Infof("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	logger.Logger.Infof("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func Init() *calendar.Service {
	ctx := context.Background()
	config := &oauth2.Config{
		ClientID:     config.Config.Google.ClientID,
		ClientSecret: config.Config.Google.ClientSecret,
		RedirectURL:  config.Config.Google.RedirectURIs[0],
		Scopes: []string{
			calendar.CalendarScope,
			calendar.CalendarEventsScope,
			calendar.CalendarEventsReadonlyScope,
			calendar.CalendarReadonlyScope,
			calendar.CalendarSettingsReadonlyScope,
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.Config.Google.AuthURI,
			TokenURL: config.Config.Google.TokenURI,
		},
	}
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		logger.Logger.Fatalf("Unable to retrieve Calendar client: %v", err)
	}
	return srv

}
func icalIdToGcalId(icalId string) string {
	idMd5 := md5.Sum([]byte(icalId))
	return hex.EncodeToString(idMd5[:])
}

func valarmTriggerToMinutes(valarmTrigger string) int64 {
	valarmTrigger = strings.ToLower(valarmTrigger)
	if strings.HasPrefix(valarmTrigger, "pt") {
		d, err := time.ParseDuration(valarmTrigger[2:])
		if err == nil {
			return int64(d.Minutes())
		}
	} else if strings.HasPrefix(valarmTrigger, "-pt") {
		d, err := time.ParseDuration(valarmTrigger[3:])
		if err == nil {
			return int64(d.Minutes())
		}
	}
	return 0
}

func icalEventToGcalEvent(event *ical.VEvent) *calendar.Event {
	eventSummary := event.GetProperty(ical.ComponentPropertySummary).Value
	startTime, startTimeErr := event.GetStartAt()
	endTime, endTimeErr := event.GetEndAt()
	if startTimeErr != nil || endTimeErr != nil {
		return nil
	}

	// Parsing ICS "VALARM" to google calendar "Minutes".
	reminderMinutes := map[int64]bool{0: true, 30: true}
	for _, alarm := range event.Alarms() {
		reminderMinutes[valarmTriggerToMinutes(alarm.GetProperty(ical.ComponentPropertyTrigger).Value)] = true
	}

	var reminders []*calendar.EventReminder
	for minutes, _ := range reminderMinutes {
		reminders = append(reminders,
			&calendar.EventReminder{
				Method:          "popup",
				Minutes:         int64(minutes),
				ForceSendFields: []string{"Minutes"},
			},
		)
	}

	return &calendar.Event{
		Summary: eventSummary,
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
		},
		Description: event.GetProperty(ical.ComponentPropertyDescription).Value,
		Id:          icalIdToGcalId(event.GetProperty(ical.ComponentPropertyUniqueId).Value),
		Reminders: &calendar.EventReminders{
			Overrides:       reminders,
			UseDefault:      false,
			ForceSendFields: []string{"UseDefault", "Overrides"},
		},
	}
}

func PushIcalEventsToGcal(calendarSrv *calendar.Service, events []*ical.VEvent) {
	for _, event := range events {
		gevent := icalEventToGcalEvent(event)
		if gevent == nil {
			continue
		}

		_, err := calendarSrv.Events.Insert(config.Config.CalendarId, gevent).Do()
		if err != nil {
			if err.(*googleapi.Error).Code == 409 { // Already Exists.
				_, err := calendarSrv.Events.Update(config.Config.CalendarId, gevent.Id, gevent).Do()
				if err != nil {
					logger.Logger.Errorf("Unable to update event: %s. %v\n", gevent.Summary, err)
					continue
				}
			} else {
				logger.Logger.Errorf("Unable to create event: %s. %v\n", gevent.Summary, err)
				continue
			}
		}
	}
	logger.Logger.Info("Finished updating events")
}
