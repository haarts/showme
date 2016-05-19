package main

import (
	"flag"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
)

var logLevel int

var showsApp []byte
var showApp []byte
var seasonApp []byte
var episodeApp []byte

func init() {
	const (
		logLevelUsage = "Set log level (0,1,2,3,4,5, higher is more logging)."
	)

	flag.IntVar(&logLevel, "log-level", int(log.ErrorLevel), logLevelUsage)
}

type commonEpisode struct {
	Number  int    `json:"number"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Image   struct {
		Medium   string `json:"medium"`
		Original string `json:"original"`
	} `json:"image"`
}

func findMatchingShow(file os.FileInfo) *TvMazeShow {
	contextLogger := log.WithField("file", file.Name())
	tvMaze := TvMazeClient{
		URLTemplate: tvMazeURLTemplate,
		logger:      contextLogger,
	}

	tvMazeShow, err := tvMaze.Find(file.Name())
	if err != nil || tvMazeShow == nil {
		contextLogger.Debug("No match")
		return nil
	}
	contextLogger.WithField("show", tvMazeShow.Name).Debug("Found match")

	return tvMazeShow
}

func unique(list []int) []int {
	unique := []int{}
	for _, item := range list {
		found := false
		for _, uniqueItem := range unique {
			if item == uniqueItem {
				found = true
			}
		}
		if !found {
			unique = append(unique, item)
		}
	}

	return unique
}

func seasons(show *TvMazeShow) []int {
	seasons := []int{}
	for _, episode := range show.Embedded.Episodes {
		seasons = append(seasons, int(episode.Season))
	}

	return unique(seasons)
}

func loadShowApp() error {
	var err error
	showApp, err = loadApp("apps/show.html")()
	return err
}

func loadSeasonApp() error {
	var err error
	seasonApp, err = loadApp("app/season.html")()
	return err
}

func loadEpisodeApp() error {
	var err error
	episodeApp, err = loadApp("app/episode.html")()
	return err
}

func loadApp(fileName string) func() ([]byte, error) {
	return func() ([]byte, error) {
		data, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.WithField("err", err).Errorf("Error opening %s", fileName)
			return nil, err
		}
		return data, nil
	}
}

func main() {
	flag.Parse()
	log.SetLevel(log.Level(logLevel))

	if err := loadShowsApp(); err != nil {
		return
	}

	if err := loadShowApp(); err != nil {
		return
	}

	if err := loadSeasonApp(); err != nil {
		return
	}

	if err := loadEpisodeApp(); err != nil {
		return
	}

	if err := os.Chdir(flag.Args()[0]); err != nil {
		log.WithFields(log.Fields{
			"err":  err,
			"root": flag.Args()[0],
		}).Fatal("Error changing working dir")
	}

	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Error initializing Fetcher")
	}

	shows := []ShowInList{}
	for _, file := range files {
		if !file.IsDir() {
			log.WithField("file", file.Name()).Warn("skipping")
			continue
		}

		tvMazeShow := findMatchingShow(file)
		if tvMazeShow != nil {
			show := convertToShowInList(tvMazeShow)
			shows = append(shows, show)

			writeShowJSON(tvMazeShow)     // 1x show.json
			writeSeasonJSONs(tvMazeShow)  // Nx season.json
			writeEpisodeJSONs(tvMazeShow) // Mx episode.json
		}
	}

	writeShows(shows)
}
