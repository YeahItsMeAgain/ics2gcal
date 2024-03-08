package config

import (
	"encoding/json"
	"log"
	"os"
)

type ConfigT struct {
	WebcalURL string
	CalendarId string
	UpdateIntervalMins  int
	Google struct {
		ClientID     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		RedirectURIs []string `json:"redirect_uris"`
		AuthURI      string   `json:"auth_uri"`
		TokenURI     string   `json:"token_uri"`
	}
}

var Config *ConfigT

func Init() {
	file, _ := os.Open("config.json")
	defer file.Close()

	decoder := json.NewDecoder(file)
	err := decoder.Decode(&Config)
	if err != nil {
		log.Fatal("[!] Can't read config.json", err)
	}
}
