package ics

import (
	"ics2gcal/logger"

	"net/http"
	"strings"

	ical "github.com/arran4/golang-ical"
)

func fetchWebcal(webcalURL string) (*http.Response, error) {
	// Replace webcal:// with https:// to follow the redirect
	url := strings.Replace(webcalURL, "webcal://", "https://", 1)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	// Check if the status code indicates a redirect
	// And follow the redirect recursively.
	if isRedirect(resp.StatusCode) {
		location := resp.Header.Get("Location")
		return fetchWebcal(location)
	}
	return resp, nil
}

func isRedirect(statusCode int) bool {
	switch statusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect:
		return true
	}
	return false
}

func ParseFromWebcal(webcalURL string) ([]*ical.VEvent, error) {
	resp, err := fetchWebcal(webcalURL)
	if err != nil {
		logger.Logger.Fatalf("Failed to fetch iCalendar file: %v", err)
	}
	defer resp.Body.Close()

	calendar, err := ical.ParseCalendar(resp.Body)
	if err != nil {
		return nil, err
	}

	events := calendar.Events()
	return events, nil
}
