package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	trakt "github.com/42minutes/go-trakt"
)

var root = "/tmp/Videos"

func listShows(root string) []os.FileInfo {
	fs, err := ioutil.ReadDir(root)
	if err != nil {
		fmt.Printf("err %+v\n", err)
	}

	var shows []os.FileInfo
	for _, fileinfo := range fs {
		if fileinfo.IsDir() {
			shows = append(shows, fileinfo)
		}
	}

	return shows
}

type Trakt struct {
	*trakt.Client
}

type season struct {
	trakt.Season
	episodes []trakt.Episode
}

type FullShow struct {
	show    trakt.Show
	seasons []season
}

func (t Trakt) turnDirsIntoShows(dirs []os.FileInfo) map[os.FileInfo]trakt.ShowResult {
	shows := make(map[os.FileInfo]trakt.ShowResult)

	for _, d := range dirs {
		results, response := t.Shows().Search(path.Base(d.Name()))
		if response.Err != nil {
			continue
		}

		shows[d] = results[0]
	}

	return shows
}

func (t Trakt) turnShowResultsIntoShows(showResults map[os.FileInfo]trakt.ShowResult) map[os.FileInfo]FullShow {
	shows := make(map[os.FileInfo]FullShow)

	for dir, show := range showResults {
		result, response := t.Shows().One(show.Show.IDs.Trakt)
		if response.Err != nil {
			continue
		}

		shows[dir] = FullShow{show: *result}
	}

	return shows
}

func (t Trakt) addSeasonsAndEpisodesToShows(shows map[os.FileInfo]FullShow) {
	for k, show := range shows {
		t.addSeasons(&show)
		t.addEpisodes(&show)
		shows[k] = show
	}
}

func (t Trakt) addSeasons(show *FullShow) {
	seasons, response := t.Seasons().All(show.show.IDs.Trakt)
	if response.Err == nil {
		for _, s := range seasons {
			show.seasons = append(show.seasons, season{Season: s}) // Wow this is really weird obmitting the package name.
		}
	}
}

func (t Trakt) addEpisodes(show *FullShow) {
	for k, season := range show.seasons {
		episodes, response := t.Episodes().AllBySeason(show.show.IDs.Trakt, season.Number)
		if response.Err == nil {
			season.episodes = episodes
		}
		show.seasons[k] = season
	}
}

func main() {
	t := Trakt{
		trakt.NewClient(
			"01045164ed603042b53acf841b590f0e7b728dbff319c8d128f8649e2427cbe9",
			trakt.TokenAuth{AccessToken: "3b6f5bdba2fa56b086712d5f3f15b4e967f99ab049a6d3a4c2e56dc9c3c90462"},
		),
	}
	dirs := listShows(root)
	searchResults := t.turnDirsIntoShows(dirs)

	shows := t.turnShowResultsIntoShows(searchResults)

	t.addSeasonsAndEpisodesToShows(shows)
}

///movies.json
///shows.json
///shows/suits.json //<= contains trakt.Show value
/////shows/suits/seasons.json //<= contains []trakt.Season
///shows/suits/seasons/1.json //<= contains trakt.Season // Does that make sense?
///shows/suits/seasons/2.json
/////shows/suits/seasons/1/episodes.json //<= contains []trakt.Episode
///shows/suits/seasons/1/episodes/1.json //<= contains trakt.Episode
///shows/suits/seasons/1/episodes/1.mp4
