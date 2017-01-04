package main

import (
	"encoding/json"
	"os"
	"path"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

type SingleShow struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Image   struct {
		Medium   string `json:"medium"`
		Original string `json:"original"`
	} `json:"image"`

	SeasonURLs []string `json:"season_urls"`
}

func writeShow(show *show) {
	writeShowJSON(show)
	writeShowApp(show.path)
}

func writeShowApp(showName string) {
	app, err := os.Create(path.Join(showName, "index.html"))
	if err != nil {
		log.WithField("err", err).Error("Error creating index.html in show root")
		return
	}
	_, err = app.Write(showApp)
	if err != nil {
		log.WithField("err", err).Error("Error writing index.html in show root")
		return
	}
}

func writeShowJSON(show *show) error {
	singleShow := SingleShow{
		Name:       show.Name,
		Summary:    show.Summary,
		Image:      show.Image,
		SeasonURLs: []string{},
	}

	for _, season := range seasons(show) {
		if _, err := os.Stat(path.Join(show.path, strconv.Itoa(season))); err == nil {
			singleShow.SeasonURLs = append(singleShow.SeasonURLs, documentRoot+show.path+"/"+strconv.Itoa(season))
		}
	}

	file, err := os.Create(path.Join(show.path, "show.json"))
	if err != nil {
		log.WithField("err", err).Warn("failed to create show.json")
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(singleShow); err != nil {
		log.WithField("err", err).Warn("failed to encode")
		return err
	}

	log.WithFields(log.Fields{
		"file": file.Name(),
	}).Info("show written to disk")

	return nil
}
