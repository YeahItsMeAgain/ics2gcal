package config

import (
	"encoding/json"
	"log"
	"os"
)

type ConfigT struct {
	WebcalURL string
	CalendarId string
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
