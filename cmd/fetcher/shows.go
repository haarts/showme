package main

import (
	"encoding/json"
	"os"

	log "github.com/Sirupsen/logrus"
)

type ShowInList struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Image   struct {
		Medium   string `json:"medium"`
		Original string `json:"original"`
	} `json:"image"`

	URL string `json:"url"`
}

func convertToShowInList(show *show) ShowInList {
	return ShowInList{
		Name:    show.Name,
		Summary: show.Summary,
		Image:   show.Image,
		URL:     "/" + show.path,
	}
}

func writeShowsJSON(shows []ShowInList) {
	file, err := os.Create("shows.json")
	if err != nil {
		log.WithField("err", err).Error("Error creating shows.json")
		return
	}

	if err = json.NewEncoder(file).Encode(shows); err != nil {
		log.WithField("err", err).Error("Error writing shows.json")
		return
	}
}

func writeShowsApp() {
	app, err := os.Create("index.html")
	if err != nil {
		log.WithField("err", err).Error("Error creating index.html in shows root")
		return
	}
	_, err = app.Write(showsApp)
	if err != nil {
		log.WithField("err", err).Error("Error writing index.html in shows root")
		return
	}
}

func writeShows(shows []ShowInList) {
	writeShowsJSON(shows)
	writeShowsApp()
}
