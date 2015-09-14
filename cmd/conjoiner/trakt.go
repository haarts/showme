package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/42minutes/go-trakt"
	log "github.com/Sirupsen/logrus"
)

type Trakt struct {
	*trakt.Client
}

type episode struct {
	trakt.Episode
	URL      string `json:"url"` // Useful when having a list of episodes and you want the single episode.
	VideoURL string `json:"video_url"`
}

type season struct {
	trakt.Season
	episodes    []episode
	URL         string `json:"url"` // Useful when season is presented in a list.
	EpisodesURL string `json:"episodes_url"`
}

type show struct {
	trakt.Show
	seasons    []season
	URL        string `json:"url"` // Useful when show is presented in a list.
	SeasonsURL string `json:"seasons_url"`
}

func (s show) findSeason(number int) (season, error) {
	for _, season := range s.seasons {
		if season.Number == number {
			return season, nil
		}
	}

	return season{}, fmt.Errorf("Could not find season %d of %s", number, s.Title)
}

func (t Trakt) turnDirsIntoShows(dirs []os.FileInfo) map[os.FileInfo]trakt.ShowResult {
	shows := make(map[os.FileInfo]trakt.ShowResult)

	for _, d := range dirs {
		log.WithFields(log.Fields{
			"dir": d,
		}).Debug("Searching for show.")
		var results []trakt.ShowResult
		var response *trakt.Result
		operation := func() error {
			showName := strings.Replace(path.Base(d.Name()), " (US)", "", 1) //RLY? Trakt is very broken.
			results, response = t.Shows().Search(showName)
			return response.Err
		}
		retry(operation)

		if len(results) > 0 {
			log.WithFields(log.Fields{
				"dir":        d,
				"show_title": results[0].Show.Title,
			}).Debug("Matched directory with show")
			shows[d] = results[0]
		}
	}

	return shows
}

func (t Trakt) turnShowResultsIntoShows(showResults map[os.FileInfo]trakt.ShowResult) map[os.FileInfo]show {
	shows := make(map[os.FileInfo]show)

	for dir, s := range showResults {
		result, response := t.Shows().One(s.Show.IDs.Trakt)
		if response.Err != nil {
			continue
		}

		shows[dir] = show{Show: *result}
	}

	return shows
}

func (t Trakt) addSeasonsAndEpisodesToShows(shows map[os.FileInfo]show) {
	for k, show := range shows {
		t.addSeasons(&show)
		t.addEpisodes(&show)
		shows[k] = show
	}
}

func (t Trakt) addSeasons(show *show) {
	seasons, response := t.Seasons().All(show.IDs.Trakt)
	if response.Err == nil {
		for _, s := range seasons {
			show.seasons = append(show.seasons, season{Season: s}) // Wow this is really weird obmitting the package name.
		}
	}
}

func (t Trakt) addEpisodes(show *show) {
	for k, season := range show.seasons {
		episodes, response := t.Episodes().AllBySeason(show.IDs.Trakt, season.Number)
		if response.Err == nil {
			for _, e := range episodes {
				season.episodes = append(season.episodes, episode{Episode: e})
			}
		}
		show.seasons[k] = season
	}
}
