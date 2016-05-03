package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

var logLevel int
var root string

func init() {
	const (
		logLevelUsage = "Set log level (0,1,2,3,4,5, higher is more logging)."
	)

	flag.IntVar(&logLevel, "log-level", int(log.ErrorLevel), logLevelUsage)
}

type ShowInList struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Image   struct {
		Medium   string `json:"medium"`
		Original string `json:"original"`
	} `json:"image"`

	URL string `json:"url"`
}

type SingleShow struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Image   struct {
		Medium   string `json:"medium"`
		Original string `json:"original"`
	} `json:"image"`

	SeasonURLs []string `json:"season_urls"`
}

type Season struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Image   struct {
		Medium   string `json:"medium"`
		Original string `json:"original"`
	} `json:"image"`

	Number   int `json:"number"`
	Episodes []struct {
		Number  int    `json:"number"`
		Name    string `json:"name"`
		Summary string `json:"summary"`
		Image   struct {
			Medium   string `json:"medium"`
			Original string `json:"original"`
		} `json:"image"`

		URL string `json:"url"`
	} `json:"episodes"`
}

type SingleEpisode struct {
	Number  int    `json:"number"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Image   struct {
		Medium   string `json:"medium"`
		Original string `json:"original"`
	} `json:"image"`

	VideoURL     string `json:"video_url"`
	ShowName     string `json:"show_name"`
	SeasonNumber int    `json:"season_number"`
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

func writeEpisodeJSONs(show *TvMazeShow) error {
	return nil
}

func writeSeasonJSONs(show *TvMazeShow) error {
	return nil
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

func writeShowJSON(show *TvMazeShow) error {
	singleShow := SingleShow{
		Name:       show.Name,
		Summary:    show.Summary,
		Image:      show.Image,
		SeasonURLs: []string{},
	}

	seasons := []int{}
	for _, episode := range show.Embedded.Episodes {
		seasons = append(seasons, int(episode.Season))
	}

	uniqueSeasons := unique(seasons)

	for _, season := range uniqueSeasons {
		if _, err := os.Stat(path.Join(show.Name, strconv.Itoa(season))); err == nil {
			singleShow.SeasonURLs = append(singleShow.SeasonURLs, "/"+show.Name+"/"+strconv.Itoa(season))
		}
	}

	file, err := os.Create(path.Join(show.Name, "show.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(singleShow); err != nil {
		return err
	}

	return nil
}

func convertToShowInList(show *TvMazeShow) ShowInList {
	return ShowInList{
		Name:    show.Name,
		Summary: show.Summary,
		Image:   show.Image,
		URL:     "/" + show.Name,
	}
}

func writeShowsJSON(shows []ShowInList) {
	file, err := os.Create("shows.json")
	if err != nil {
		log.WithField("err", err).Error("Error opening shows.json")
		return
	}

	if err = json.NewEncoder(file).Encode(shows); err != nil {
		log.WithField("err", err).Error("Error writing shows.json")
		return
	}
}

func main() {
	flag.Parse()
	log.SetLevel(log.Level(logLevel))
	log.Info("Started fetcher")

	root = flag.Args()[0]

	if err := os.Chdir(root); err != nil {
		log.WithFields(log.Fields{
			"err":  err,
			"root": root,
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
			log.WithField("file", file.Name()).Debug("skipping")
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

	writeShowsJSON(shows)
}
