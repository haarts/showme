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

// TODO deal with broken symlinks

type conjoiner struct {
	root                 string
	isShowsRootRegexp    *regexp.Regexp
	isSeasonsRootRegexp  *regexp.Regexp
	isEpisodesRootRegexp *regexp.Regexp
}

func newConjoiner(root string) (*conjoiner, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	trailingName := string(filepath.Separator) + "[^" + string(filepath.Separator) + "]+"

	showsRoot := root + trailingName
	seasonsRoot := showsRoot + trailingName
	episodesRoot := seasonsRoot + trailingName

	return &conjoiner{
		root:                 root,
		isShowsRootRegexp:    regexp.MustCompile(showsRoot + "\\z"),
		isSeasonsRootRegexp:  regexp.MustCompile(seasonsRoot + "\\z"),
		isEpisodesRootRegexp: regexp.MustCompile(episodesRoot + "\\z"),
	}, nil
}

func (c conjoiner) isShowRoot(dir string) (bool, error) {
	f, err := os.Stat(dir)
	if err != nil {
		return false, err
	}

	return c.isShowsRootRegexp.MatchString(dir) && f.IsDir(), nil
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
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Error occured when listing shows")
		return []os.FileInfo{}
	}

	var shows []os.FileInfo
	for _, fileinfo := range fs {
		if fileinfo.IsDir() {
			shows = append(shows, fileinfo)
		}
	}

	return shows
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
	c, err := newConjoiner(os.Args[1])
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Error initializing Conjoiner")
	}

	shows := c.lookup()
	log.WithFields(log.Fields{
		"#shows": len(shows),
	}).Info("Found shows")

	err = c.createJSONs(shows)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("An error occurred while writing JSON files")
	}
}
