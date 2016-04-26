// fetcher takes a directory with series and checks which episodes are already
// downloaded. For this it first tries to match the directory name to a search
// result from tvmaze.
package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
)

var logLevel int
var root string

func init() {
	const (
		logLevelUsage = "Set log level (0,1,2,3,4,5, higher is more logging)."
	)

	flag.IntVar(&logLevel, "log-level", int(log.ErrorLevel), logLevelUsage)
}

type Show struct {
	TvMazeShow
	Seasons Seasons
}

type Season struct {
	Number int64 `json:"number"`
}

type Seasons []Season

func (s *Seasons) Add(season Season) {
	for _, existing := range *s {
		if existing.Number == season.Number {
			return
		}
	}
	*s = append(*s, season)
}

func findMatchingShow(file os.FileInfo) *Show {
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
	spew.Dump(tvMazeShow)

	show := mapFoundShowToDiskContent(file, tvMazeShow)

	contextLogger.WithField("show", show.Name).Debug("Found match")
	return &show
}

func mapFoundShowToDiskContent(file os.FileInfo, tvMazeShow *TvMazeShow) Show {
	contextLogger := log.WithField("file", file.Name())
	seasons := Seasons{}

	for _, episode := range tvMazeShow.Embedded.Episodes {
		//seasonDir, err := os.Stat(path.Join(root, file.Name(), strconv.Itoa(int(episode.Season))))
		season := strconv.Itoa(int(episode.Season))
		_, err := os.Stat(path.Join(root, file.Name(), season))
		//fmt.Printf("seasonDir = %+v\n", seasonDir)
		if err == nil {
			seasons.Add(Season{Number: episode.Season})
		} else if os.IsNotExist(err) {
			contextLogger.WithField("season", season).Info("Season not found on disk")
		} else {
			contextLogger.WithField("err", err).Error("Failed to stat season dir")
		}
	}

	//spew.Dump(seasons)
	return Show{TvMazeShow: *tvMazeShow, Seasons: seasons}
}

func main() {
	flag.Parse()
	log.SetLevel(log.Level(logLevel))
	log.Info("Started fetcher")

	root = flag.Args()[0]

	files, err := ioutil.ReadDir(root)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Error initializing Fetcher")
	}

	for _, file := range files {
		if file.IsDir() { // TODO make parralel
			findMatchingShow(file)
		} else {
			contextLogger := log.WithField("file", file.Name())
			contextLogger.Debug("skipping")
		}
	}
}
