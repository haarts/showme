package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"

	trakt "github.com/42minutes/go-trakt"
)

type conjoiner struct {
	root                 string
	isShowRootRegexp     *regexp.Regexp
	isSeasonsRootRegexp  *regexp.Regexp
	isEpisodesRootRegexp *regexp.Regexp
}

func newConjoiner(root string) *conjoiner {
	trailingName := string(filepath.Separator) + "[^" + string(filepath.Separator) + "]+"

	showRoot := root + trailingName
	seasonsRoot := showRoot + trailingName
	episodesRoot := seasonsRoot + trailingName

	return &conjoiner{
		root:                 root,
		isShowRootRegexp:     regexp.MustCompile(showRoot + "\\z"),
		isSeasonsRootRegexp:  regexp.MustCompile(seasonsRoot + "\\z"),
		isEpisodesRootRegexp: regexp.MustCompile(episodesRoot + "\\z"),
	}
}

func (c conjoiner) isShowRoot(dir string) bool {
	f, _ := os.Stat(dir)
	return c.isShowRootRegexp.MatchString(dir) && f.IsDir()
}

func (c conjoiner) isSeasonsRoot(dir string) bool {
	f, _ := os.Stat(dir)
	return c.isSeasonsRootRegexp.MatchString(dir) && f.IsDir()
}

func (c conjoiner) listShows() []os.FileInfo {
	fs, err := ioutil.ReadDir(c.root)
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

func (c conjoiner) lookup() map[os.FileInfo]FullShow {
	t := Trakt{
		trakt.NewClient(
			"01045164ed603042b53acf841b590f0e7b728dbff319c8d128f8649e2427cbe9",
			trakt.TokenAuth{AccessToken: "3b6f5bdba2fa56b086712d5f3f15b4e967f99ab049a6d3a4c2e56dc9c3c90462"},
		),
	}
	dirs := c.listShows()
	searchResults := t.turnDirsIntoShows(dirs)

	shows := t.turnShowResultsIntoShows(searchResults)

	t.addSeasonsAndEpisodesToShows(shows)

	return shows
}

func writeObject(v interface{}, dir string) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dir, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (s FullShow) findSeason(number int) (season, error) {
	for _, season := range s.seasons {
		if season.Number == number {
			return season, nil
		}
	}

	return season{}, fmt.Errorf("Could not find season %d", number)
}

func (c conjoiner) showFunc(show FullShow) filepath.WalkFunc {
	return func(dir string, info os.FileInfo, err error) error {
		if c.isShowRoot(dir) {
			err = writeObject(show.seasons, path.Join(dir, "seasons.json"))
			if err != nil {
				return err
			}
			for _, season := range show.seasons {
				err := writeObject(season, path.Join(dir, strconv.Itoa(season.Number)+".json"))
				if err != nil {
					return err
				}
			}
		}
		if c.isSeasonsRoot(dir) {
			_, seasonNumber := filepath.Split(dir)
			i, err := strconv.Atoi(seasonNumber)
			if err != nil {
				return err
			}
			season, err := show.findSeason(i)
			if err != nil {
				return err
			}
			err = writeObject(season.episodes, path.Join(dir, "episodes.json"))

			for _, episode := range season.episodes {
				err = writeObject(episode, path.Join(dir, episode.Title+".json"))
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func (c conjoiner) createJSONs(shows map[os.FileInfo]FullShow) error {
	for dir, show := range shows {
		err := filepath.Walk(dir.Name(), c.showFunc(show))
		if err != nil {
			return err
		}
	}

	var showsIndex []trakt.Show
	for dir, show := range shows {
		showsIndex = append(showsIndex, show.show)
		err := writeObject(show.show, path.Join(dir.Name(), "..", show.show.Title+".json"))
		if err != nil {
			return err
		}
	}

	err := writeObject(showsIndex, path.Join(c.root, "shows.json"))
	if err != nil {
		return err
	}

	return nil
}

func main() {
	c := newConjoiner("/tmp/Videos")
	shows := c.lookup()
	fmt.Printf("shows %+v\n", shows)

	//writeShowIndex(shows)
	//writeIndividualShows(shows)
	//writeSeasonIndex(shows)
	//writeIndividualSeasons(show)
	//writeEpisodeIndex(show)
	//writeIndividualEpisodes(show)
}
