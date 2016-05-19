package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type SingleEpisode struct {
	commonEpisode

	VideoURL     string `json:"video_url"`
	ShowName     string `json:"show_name"`
	SeasonNumber int    `json:"season_number"`
}

func episodeExists(seasonDir string, episode TvMazeEpisode) bool {
	if episodeVideoFile(seasonDir, episode) == "" {
		return false
	}
	return true
}

func urlify(name string) string {
	re := regexp.MustCompile("[^a-zA-Z0-9]")
	return re.ReplaceAllString(name, "-")
}

func writeEpisodeJSON(episode SingleEpisode) error {
	episodeDir := path.Join(
		episode.ShowName,
		strconv.Itoa(episode.SeasonNumber),
		urlify(episode.Name),
	)

	if _, err := os.Stat(episodeDir); err != nil {
		err := os.Mkdir(episodeDir, 0744)
		if err != nil {
			return err
		}
	}

	file, err := os.Create(path.Join(
		episodeDir,
		"episode.json"),
	)
	if err != nil {
		log.WithField("err", err).Warn("failed to create episode.json")
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(episode); err != nil {
		log.WithField("err", err).Warn("failed to encode to episode.json")
		return err
	}

	log.WithFields(log.Fields{
		"file": file.Name(),
	}).Debug("episode written to disk")

	return nil
}

func episodes(seasonNumber int, show *TvMazeShow) []SingleEpisode {
	episodes := []SingleEpisode{}

	for _, episode := range show.Embedded.Episodes {
		if int(episode.Season) != seasonNumber {
			log.WithFields(log.Fields{
				"actual_season":   episode.Season,
				"required_season": seasonNumber,
			}).Debug("wrong season, skipping")
			continue
		}

		// Check if episode exists on disk
		if !episodeExists(path.Join(show.Name, strconv.Itoa(seasonNumber)), episode) {
			log.WithFields(log.Fields{
				"episode": episode.Episode,
				"name":    episode.Name,
				"path":    path.Join(show.Name, strconv.Itoa(seasonNumber)),
			}).Warn("episode doesn't exists on disk, skipping")
			continue
		}

		episodes = append(episodes, SingleEpisode{
			commonEpisode: commonEpisode{
				Number:  int(episode.Episode),
				Name:    episode.Name,
				Summary: episode.Summary,
				Image:   episode.Image,
			},

			VideoURL: path.Join(show.Name, strconv.Itoa(seasonNumber), episodeVideoFile(
				path.Join(show.Name, strconv.Itoa(seasonNumber)), episode)),
			ShowName:     show.Name,
			SeasonNumber: seasonNumber,
		})
	}

	return episodes
}

func writeEpisodeJSONs(show *TvMazeShow) {
	for _, seasonNumber := range seasons(show) {
		if _, err := os.Stat(path.Join(show.Name, strconv.Itoa(seasonNumber))); err != nil {
			log.WithFields(log.Fields{
				"err":    err,
				"season": seasonNumber,
				"show":   show.Name,
			}).Warn("season not found on disk, skipping")
			continue
		}

		for _, episode := range episodes(seasonNumber, show) {
			if err := writeEpisodeJSON(episode); err != nil {
				log.WithFields(log.Fields{
					"err":     err,
					"season":  seasonNumber,
					"episode": episode.Number,
				}).Error("Error writing episode")
			}
		}
	}
}

func episodeVideoFile(seasonDir string, episode TvMazeEpisode) string {
	files, err := ioutil.ReadDir(seasonDir)
	if err != nil {
		log.WithFields(log.Fields{
			"season": episode.Season,
			"err":    err,
		}).Error("Error reading season directory")
	}

	for _, file := range files {
		if file.IsDir() {
			log.WithField("file", file.Name()).Debug("is dir, skipping")
			continue
		}

		if strings.Contains(file.Name(), fmt.Sprintf("S%02dE%02d", episode.Season, episode.Episode)) {
			log.WithField("file", file.Name()).Debug("match")
			return file.Name()
		}
	}

	return ""
}
