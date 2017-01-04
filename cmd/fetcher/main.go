package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/xrash/smetrics"
)

var logLevel int
var documentRoot string

var showsApp []byte
var showApp []byte
var seasonApp []byte
var episodeApp []byte

func init() {
	const (
		logLevelUsage = "Set log level (0,1,2,3,4,5, higher is more logging)."
		documentRootUsage = "Set the document root of the URLs in the to be generated JSON files."
	)

	flag.IntVar(&logLevel, "log-level", int(log.ErrorLevel), logLevelUsage)
	flag.StringVar(&documentRoot, "document-root", "/", documentRootUsage)
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

type show struct {
	TvMazeShow
	path string
}

func findMatchingShow(filename string) *show {
	contextLogger := log.WithField("file", filename)
	tvMaze := TvMazeClient{
		logger: contextLogger,
	}

	tvMazeShow, err := tvMaze.Find(filename)
	if err != nil || tvMazeShow == nil || !goodEnoughMatch(filename, tvMazeShow.Name) {
		contextLogger.Debug("No match")
		return nil
	}
	contextLogger.WithField("show", tvMazeShow.Name).Debug("Found match")

	return &show{
		TvMazeShow: *tvMazeShow,
		path:       filename,
	}
}

func goodEnoughMatch(s1, s2 string) bool {
	if smetrics.JaroWinkler(s1, s2, 0.7, 8) < 0.95 {
		return false
	}
	return true
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

func seasons(show *show) []int {
	seasons := []int{}
	for _, episode := range show.Embedded.Episodes {
		seasons = append(seasons, int(episode.Season))
	}

	return unique(seasons)
}

func loadShowsApp() error {
	var err error
	showsApp, err = loadApp("apps/shows.html")()
	return err
}

func loadShowApp() error {
	var err error
	showApp, err = loadApp("apps/show.html")()
	return err
}

func loadSeasonApp() error {
	var err error
	seasonApp, err = loadApp("apps/season.html")()
	return err
}

func loadEpisodeApp() error {
	var err error
	episodeApp, err = loadApp("apps/episode.html")()
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

	dir, err := os.Getwd()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Error getting working directory")
	}

	if len(flag.Args()) != 1 {
		log.Fatal("Require one argument pointing to media path")
	}

	if err := os.Chdir(path.Join(dir, flag.Args()[0])); err != nil {
		log.WithFields(log.Fields{
			"err":  err,
			"root": path.Join(dir, flag.Args()[0]),
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

		show := findMatchingShow(file.Name())
		if show != nil {
			showInList := convertToShowInList(show)
			shows = append(shows, showInList)

			writeShow(show)     // 1x show.json
			writeSeasons(show)  // Nx season.json
			writeEpisodes(show) // Mx episode.json
		}
	}

	writeShows(shows)
}
