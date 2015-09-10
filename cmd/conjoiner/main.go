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
	"strings"

	trakt "github.com/42minutes/go-trakt"
	log "github.com/Sirupsen/logrus"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

type conjoiner struct {
	root                 string
	isShowRootRegexp     *regexp.Regexp
	isSeasonsRootRegexp  *regexp.Regexp
	isEpisodesRootRegexp *regexp.Regexp
}

func newConjoiner(root string) *conjoiner {
	trailingName := string(filepath.Separator) + "[^" + string(filepath.Separator) + "]+"

	showRoot := filepath.Base(root) + trailingName
	seasonsRoot := showRoot + trailingName
	episodesRoot := seasonsRoot + trailingName

	return &conjoiner{
		root:                 root,
		isShowRootRegexp:     regexp.MustCompile(showRoot + "\\z"),
		isSeasonsRootRegexp:  regexp.MustCompile(seasonsRoot + "\\z"),
		isEpisodesRootRegexp: regexp.MustCompile(episodesRoot + "\\z"),
	}
}

func (c conjoiner) isShowRoot(dir string) (bool, error) {
	f, err := os.Stat(dir)
	if err != nil {
		return false, err
	}

	return c.isShowRootRegexp.MatchString(dir) && f.IsDir(), nil
}

func (c conjoiner) isSeasonsRoot(dir string) (bool, error) {
	f, err := os.Stat(dir)
	if err != nil {
		return false, err
	}

	return c.isSeasonsRootRegexp.MatchString(dir) && f.IsDir(), nil
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

func retry(f func() error) error {
	var err error
	for i := 0; i < 3; i++ {
		if err = f(); err == nil {
			break
		}
	}

	return err
}

func (t Trakt) turnDirsIntoShows(dirs []os.FileInfo) map[os.FileInfo]trakt.ShowResult {
	shows := make(map[os.FileInfo]trakt.ShowResult)

	for _, d := range dirs {
		var results []trakt.ShowResult
		var response *trakt.Result
		operation := func() error {
			showName := strings.Replace(path.Base(d.Name()), " (US)", "", 1) //RLY? Trakt is very broken.
			results, response = t.Shows().Search(showName)
			return response.Err
		}
		retry(operation)

		if len(results) > 0 {
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

func (c conjoiner) lookup() map[os.FileInfo]show {
	t := Trakt{
		trakt.NewClientWith(
			"https://api-v2launch.trakt.tv",
			trakt.UserAgent,
			"01045164ed603042b53acf841b590f0e7b728dbff319c8d128f8649e2427cbe9",
			trakt.TokenAuth{AccessToken: "3b6f5bdba2fa56b086712d5f3f15b4e967f99ab049a6d3a4c2e56dc9c3c90462"},
			nil,
		),
	}
	dirs := c.listShows()
	searchResults := t.turnDirsIntoShows(dirs)

	shows := t.turnShowResultsIntoShows(searchResults)

	t.addSeasonsAndEpisodesToShows(shows)

	return shows
}

func writeObject(v interface{}, file string) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (s show) findSeason(number int) (season, error) {
	for _, season := range s.seasons {
		if season.Number == number {
			return season, nil
		}
	}

	return season{}, fmt.Errorf("Could not find season %d of %s", number, s.Title)
}

func withoutRoot(root, path string) string {
	return strings.Replace(path, root+string(filepath.Separator), "", 1)
}

func (c conjoiner) showFunc(show show) filepath.WalkFunc {
	return func(dir string, info os.FileInfo, err error) error {
		isShowRoot, err := c.isShowRoot(dir)
		if err != nil {
			return err
		}

		if isShowRoot {
			for i, season := range show.seasons {
				location := path.Join(dir, strconv.Itoa(season.Number)+".json")
				show.seasons[i].URL = withoutRoot(c.root, location)
				show.seasons[i].EpisodesURL =
					withoutRoot(c.root, path.Join(dir, strconv.Itoa(season.Number), "episodes.json"))
				err := writeObject(show.seasons[i], location) // write single season JSON
				if err != nil {
					return err
				}
			}

			err = writeObject(show.seasons, path.Join(dir, "seasons.json")) // write seasons as a list
			if err != nil {
				return err
			}
		}

		isSeasonsRoot, err := c.isSeasonsRoot(dir)
		if err != nil {
			return err
		}

		if isSeasonsRoot {
			_, seasonNumber := filepath.Split(dir)
			i, err := strconv.Atoi(seasonNumber)
			if err != nil {
				return err
			}
			season, err := show.findSeason(i)
			if err != nil {
				return err
			}

			for i, episode := range season.episodes {
				videoLocation, err := matchNameWithVideo(episode, dir)
				if err == nil {
					episode.VideoURL = withoutRoot(c.root, path.Join(dir, videoLocation))
				}

				location := path.Join(
					dir,
					fmt.Sprintf("s%02de%02d %s.json", episode.Season, episode.Number, replaceSeperators(episode.Title)),
				)
				episode.URL = withoutRoot(c.root, location)

				err = writeObject(episode, location) // write single episode JSON
				if err != nil {
					return err
				}
				season.episodes[i] = episode
			}

			err = writeObject(season.episodes, path.Join(dir, "episodes.json")) // write episodes as a list
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func replaceSeperators(name string) string {
	re := regexp.MustCompile(string(filepath.Separator))
	return string(re.ReplaceAll([]byte(name), []byte(" ")))
}

func matchNameWithVideo(episode episode, dir string) (string, error) {
	asRunes := []rune(episode.Title)
	var best string
	var bestScore = 999
	commonNotation := fmt.Sprintf("s%02de%02d", episode.Season, episode.Number)

	fs, _ := ioutil.ReadDir(dir)
	for _, f := range fs {
		b, _ := regexp.MatchString(`\.(mp4)\z`, f.Name())
		if !b {
			continue
		}

		// Bail out early
		if ok, _ := regexp.Match(commonNotation, []byte(f.Name())); ok {
			return f.Name(), nil
		}

		score := levenshtein.DistanceForStrings(asRunes, []rune(f.Name()), levenshtein.DefaultOptions)
		if score < bestScore {
			bestScore = score
			best = f.Name()
		}
	}

	if bestScore > 15 { // too bad to consider
		return "", fmt.Errorf("no match found")
	}

	return path.Join(dir, best), nil
}

func (c conjoiner) createJSONs(shows map[os.FileInfo]show) error {
	for dir, show := range shows {
		err := filepath.Walk(path.Join(c.root, dir.Name()), c.showFunc(show))
		if err != nil {
			return err
		}
	}

	var showIndex []show
	for _, show := range shows {
		URL := show.Title + ".json"
		show.URL = URL
		show.SeasonsURL = path.Join(show.Title, "seasons.json")

		err := writeObject(show, path.Join(c.root, URL)) // write single show JSON
		if err != nil {
			return err
		}
		showIndex = append(showIndex, show)
	}

	err := writeObject(showIndex, path.Join(c.root, "shows.json")) // write shows as a list
	if err != nil {
		return err
	}

	return nil
}

func main() {
	log.Info("Started conjoiner")
	c := newConjoiner(os.Args[1])

	shows := c.lookup()
	log.WithFields(log.Fields{
		"#shows": len(shows),
	}).Info("Found shows")

	err := c.createJSONs(shows)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("An error occurred while writing JSON files")
	}
}
