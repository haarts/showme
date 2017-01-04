package main

import (
	"encoding/json"
	"os"
	"path"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

type Season struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Image   struct {
		Medium   string `json:"medium"`
		Original string `json:"original"`
	} `json:"image"`

	Number   int               `json:"number"`
	Episodes []internalEpisode `json:"episodes"`
}

type internalEpisode struct {
	commonEpisode

	URL string `json:"url"`
}

func writeSeasons(show *show) {
	for _, seasonNumber := range seasons(show) {
		if _, err := os.Stat(path.Join(show.path, strconv.Itoa(seasonNumber))); err != nil {
			continue
		}

		writeSeasonJSON(seasonNumber, show)
		writeSeasonApp(show.path, strconv.Itoa(seasonNumber))
	}
}

func writeSeasonApp(showPath, seasonNumber string) {
	app, err := os.Create(path.Join(showPath, seasonNumber, "index.html"))
	if err != nil {
		log.WithField("err", err).Error("Error creating index.html in show root")
		return
	}
	_, err = app.Write(seasonApp)
	if err != nil {
		log.WithField("err", err).Error("Error writing index.html in show root")
		return
	}
}

func writeSeasonJSON(seasonNumber int, show *show) {
	file, err := os.Create(path.Join(show.path, strconv.Itoa(seasonNumber), "season.json"))
	if err != nil {
		log.WithField("err", err).Warn("failed to create show.json")
		return
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(season(seasonNumber, show)); err != nil {
		log.WithField("err", err).Warn("failed to encode")
		return
	}

	log.WithFields(log.Fields{
		"file": file.Name(),
	}).Info("season written to disk")
}

func season(number int, show *show) Season {
	season := Season{
		Name:    show.Name,
		Summary: show.Summary,
		Image:   show.Image,
		Number:  number,
	}

	episodes := []internalEpisode{}

	for _, episode := range show.Embedded.Episodes {
		if int(episode.Season) != number {
			continue
		}

		// Check if episode exists on disk
		if !episodeExists(path.Join(show.path, strconv.Itoa(number)), episode) {
			continue
		}

		episodes = append(episodes, internalEpisode{
			commonEpisode: commonEpisode{
				Number:  int(episode.Episode),
				Name:    episode.Name,
				Summary: episode.Summary,
				Image:   episode.Image,
			},
			URL: documentRoot + path.Join(show.path, strconv.Itoa(number), urlify(episode.Name)),
		})
	}

	season.Episodes = episodes

	return season
}
