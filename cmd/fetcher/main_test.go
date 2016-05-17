package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var tvMazeShow = &TvMazeShow{
	Name: "show1",
	Embedded: struct {
		Episodes []TvMazeEpisode `json:"episodes"`
	}{
		Episodes: []TvMazeEpisode{
			TvMazeEpisode{
				Name:    "first",
				Episode: int64(1),
				Season:  int64(1), // fine, exists on disk
			},
			TvMazeEpisode{
				Name:    "second",
				Episode: int64(2),
				Season:  int64(1), // fine, exists on disk
			},
			TvMazeEpisode{
				Name:    "third",
				Episode: int64(3),
				Season:  int64(1), // not fine, absent on disk
			},
			TvMazeEpisode{
				Name:    "first in second",
				Episode: int64(1),
				Season:  int64(2), // not fine, absent on disk
			},
			TvMazeEpisode{
				Name:    "first in second",
				Episode: int64(1),
				Season:  int64(3), // not fine, absent on disk
			},
		},
	},
}

func TestConvertToShowInList(t *testing.T) {
	tvMazeShow := &TvMazeShow{
		Name:    "foo",
		Summary: "bar",
		Image: struct {
			Medium   string `json:"medium"`
			Original string `json:"original"`
		}{
			Medium:   "baz",
			Original: "buzz",
		},
	}

	show := convertToShowInList(tvMazeShow)

	assert.Equal(t, show.URL, "/foo")
	assert.Equal(t, show.Name, "foo")
	assert.NotNil(t, show.Image)
}

func TestCreateShowsJSON(t *testing.T) {
	// step 1; land on home
	// name; shows.json
	// url; https://foo.bar
	copyR("testdata/Videos_template", "testdata/Videos")
	defer os.RemoveAll("testdata/Videos")

	expected := `[
		{
			name: "foo",
			summary: "lorum ipsum",
			image: {
				medium: "https://placehold.it/350x150",
				original: "https://placehold.it/450x150",
			},
			url: "/foo"
		},
		{
			name: "bar",
			summary: "lorum ipsum",
			image: {
				medium: "https://placehold.it/350x150",
				original: "https://placehold.it/450x150",
			},
			url: "/bar"
		}
	]`

	assert.NotNil(t, expected)
}

func TestCreateShowJSON(t *testing.T) {
	// step 2; having clicked on A show
	// name; show.json
	// url; https://foo.bar/foo
	expected := `{
		name: "foo",
		summary: "lorum ipsum",
		image: {
			medium: "https://placehold.it/350x150",
			original: "https://placehold.it/450x150",
		},
		season_urls: [
			"/foo/1",
			"/foo/2",
			"/foo/3",
		]
	}`
	assert.NotNil(t, expected)

	copyR("testdata/Videos_template", "testdata/Videos")
	defer os.RemoveAll("testdata/Videos")

	require.NoError(t, os.Chdir("testdata/Videos"))
	defer os.Chdir("../..")

	require.NoError(t, writeShowJSON(tvMazeShow))

	file, err := os.Open("show1/show.json")
	require.NoError(t, err)

	show := &SingleShow{}
	require.NoError(t, json.NewDecoder(file).Decode(show))

	assert.Len(t, show.SeasonURLs, 2)
	assert.Equal(t, "/show1/1", show.SeasonURLs[0])
	assert.Equal(t, "/show1/2", show.SeasonURLs[1])
}

func TestCreateSeasonJSON(t *testing.T) {
	// step 3; having click on A season
	// name;  season.json
	// url; https://foo.bar/foo/1
	expected := `{
		name: "foo",
		summary: "lorum ipsum",
		image: {
			medium: "https://placehold.it/350x150",
			original: "https://placehold.it/450x150",
		},
		number: 1,
		episodes: [
			{
				number: 1,
				name: "pilot",
				summary: "lorum ipsum shizzle",
				image: {
					medium: "https://placehold.it/350x150",
					original: "https://placehold.it/450x150",
				},
				url: "/foo/1/pilot"
			},
			{
				number: 2,
				name: "success",
				summary: "lorum ipsum very shizzle",
				image: {
					medium: "https://placehold.it/350x150",
					original: "https://placehold.it/450x150",
				},
				url: "/foo/1/success"
			}
		]
	}`
	assert.NotNil(t, expected)

	copyR("testdata/Videos_template", "testdata/Videos")
	defer os.RemoveAll("testdata/Videos")

	require.NoError(t, os.Chdir("testdata/Videos"))
	defer os.Chdir("../..")

	writeSeasonJSONs(tvMazeShow)

	file, err := os.Open("show1/1/season.json")
	require.NoError(t, err)

	season := &Season{}
	require.NoError(t, json.NewDecoder(file).Decode(season))

	assert.Equal(t, tvMazeShow.Name, season.Name)
	assert.Equal(t, tvMazeShow.Summary, season.Summary)
	assert.Equal(t, tvMazeShow.Image, season.Image)
	assert.Equal(t, tvMazeShow.Embedded.Episodes[0].Season, season.Number)
	assert.Len(t, season.Episodes, 2)
	assert.Equal(t, "/show1/1/first", season.Episodes[0].URL)
	assert.Equal(t, "/show1/1/second", season.Episodes[1].URL)
}

func TestCreateEpisodeJSON(t *testing.T) {
	// step 4; having click on A episode
	// name; episode.json
	// url; https://foo.bar/foo/1/pilot

	expected := `{
		show_name: "foo",
		season_number: 1,
		number: 2,
		name: "success",
		summary: "lorum ipsum very shizzle",
		image: {
			medium: "https://placehold.it/350x150",
			original: "https://placehold.it/450x150",
		},
		video_url: "/foo/1/success/foo-success-S01E02.webm"
	}`

	assert.NotNil(t, expected)

	copyR("testdata/Videos_template", "testdata/Videos")
	defer os.RemoveAll("testdata/Videos")

	require.NoError(t, os.Chdir("testdata/Videos"))
	defer os.Chdir("../..")

	writeEpisodeJSONs(tvMazeShow)

	file, err := os.Open("show1/1/first/episode.json")
	require.NoError(t, err)

	episode := &SingleEpisode{}
	require.NoError(t, json.NewDecoder(file).Decode(episode))

	assert.Equal(t, tvMazeShow.Name, episode.ShowName)
	assert.Equal(t, tvMazeShow.Embedded.Episodes[0].Season, episode.SeasonNumber)
	assert.Equal(t, tvMazeShow.Embedded.Episodes[0].Name, episode.Name)
	assert.Equal(t, "show1/1/S01E01_bar.mp4", episode.VideoURL)
}

func copyR(src, dest string) {
	filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		target := strings.Replace(path, src, dest, -1)
		if info.IsDir() {
			os.MkdirAll(target, 0755)
		} else {
			copyFile(path, target)
		}
		return nil
	})
}

func copyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}

	}

	return
}
