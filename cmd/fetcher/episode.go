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

func writeEpisodes(show *show) {
	for _, seasonNumber := range seasons(show) {
		if _, err := os.Stat(path.Join(show.path, strconv.Itoa(seasonNumber))); err != nil {
			log.WithFields(log.Fields{
				"err":    err,
				"season": seasonNumber,
				"show":   show.Name,
				"path":   show.path,
			}).Warn("season not found on disk, skipping")
			continue
		}

		for _, episode := range episodes(seasonNumber, show) {
			writeEpisodeJSON(show.path, episode)
			writeEpisodeApp(show.path, episode)
		}
	}
}

func writeEpisodeApp(rootPath string, episode SingleEpisode) {
	episodeDir := path.Join(
		rootPath,
		strconv.Itoa(episode.SeasonNumber),
		urlify(episode.Name),
	)

	app, err := os.Create(path.Join(episodeDir, "index.html"))
	if err != nil {
		log.WithField("err", err).Error("Error creating index.html in shows root")
		return
	}
	_, err = app.Write(episodeApp)
	if err != nil {
		log.WithField("err", err).Error("Error writing index.html in shows root")
		return
	}
}

func writeEpisodeJSON(rootPath string, episode SingleEpisode) {
	episodeDir := path.Join(
		rootPath,
		strconv.Itoa(episode.SeasonNumber),
		urlify(episode.Name),
	)

	if _, err := os.Stat(episodeDir); err != nil {
		err := os.Mkdir(episodeDir, 0755)
		if err != nil {
			log.WithField("err", err).Error("failed to create episode directory")
			return
		}
	}

	file, err := os.Create(path.Join(
		episodeDir,
		"episode.json"),
	)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"dir": episodeDir,
		}).Warn("failed to create episode.json")
		return
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(episode); err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"dir": episodeDir,
		}).Warn("failed to encode to episode.json")
		return
	}

	log.WithFields(log.Fields{
		"file": file.Name(),
	}).Debug("episode written to disk")
}

func episodes(seasonNumber int, show *show) []SingleEpisode {
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
			}).Warn("episode doesn't exists on disk or has the wrong format, skipping")
			continue
		}

		episodes = append(episodes, SingleEpisode{
			commonEpisode: commonEpisode{
				Number:  int(episode.Episode),
				Name:    episode.Name,
				Summary: episode.Summary,
				Image:   episode.Image,
			},

			VideoURL: path.Join(show.path, strconv.Itoa(seasonNumber), episodeVideoFile(
				path.Join(show.path, strconv.Itoa(seasonNumber)), episode)),
			ShowName:     show.Name,
			SeasonNumber: seasonNumber,
		})
	}

	return episodes
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
			log.WithField("file", file.Name()).Debug("looking for video file found dir, skipping")
			continue
		}

		if strings.Contains(file.Name(), fmt.Sprintf("S%02dE%02d", episode.Season, episode.Episode)) &&
			strings.HasSuffix(file.Name(), "webm") {
			log.WithFields(log.Fields{
				"file":    file.Name(),
				"episode": episode.Episode,
				"season":  episode.Season,
			}).Debug("matched video file with episode")
			return file.Name()
		}
	}

	return ""
}
