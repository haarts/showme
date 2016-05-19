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

func writeShowJSON(show *TvMazeShow) error {
	singleShow := SingleShow{
		Name:       show.Name,
		Summary:    show.Summary,
		Image:      show.Image,
		SeasonURLs: []string{},
	}

	for _, season := range seasons(show) {
		if _, err := os.Stat(path.Join(show.Name, strconv.Itoa(season))); err == nil {
			singleShow.SeasonURLs = append(singleShow.SeasonURLs, "/"+show.Name+"/"+strconv.Itoa(season))
		}
	}

	file, err := os.Create(path.Join(show.Name, "show.json"))
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
