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

func writeSeasonJSON(seasonNumber int, show *TvMazeShow) error {
	file, err := os.Create(path.Join(show.Name, strconv.Itoa(seasonNumber), "season.json"))
	if err != nil {
		log.WithField("err", err).Warn("failed to create show.json")
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(season(seasonNumber, show)); err != nil {
		log.WithField("err", err).Warn("failed to encode")
		return err
	}

	log.WithFields(log.Fields{
		"file": file.Name(),
	}).Info("season written to disk")

	return nil
}

func writeSeasonJSONs(show *TvMazeShow) {
	for _, seasonNumber := range seasons(show) {
		if _, err := os.Stat(path.Join(show.Name, strconv.Itoa(seasonNumber))); err != nil {
			continue
		}

		// TODO perhaps I can drop it on the floor here? Eg not return anything in writeSeasonJSON
		if err := writeSeasonJSON(seasonNumber, show); err != nil {
			log.WithFields(log.Fields{
				"err":    err,
				"season": seasonNumber,
			}).Error("Error writing season")
		}
	}
}

func season(number int, show *TvMazeShow) Season {
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
		if !episodeExists(path.Join(show.Name, strconv.Itoa(number)), episode) {
			continue
		}

		episodes = append(episodes, internalEpisode{
			commonEpisode: commonEpisode{
				Number:  int(episode.Episode),
				Name:    episode.Name,
				Summary: episode.Summary,
				Image:   episode.Image,
			},
			URL: "/" + path.Join(show.Name, strconv.Itoa(number), urlify(episode.Name)),
		})
	}

	season.Episodes = episodes

	return season
}
