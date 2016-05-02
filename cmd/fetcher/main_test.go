package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	copyR("testdata/Videos_template", "testdata/Videos")
	defer os.RemoveAll("testdata/Videos")

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

	tvMazeShow := &TvMazeShow{
		Name: "show1",
		Embedded: struct {
			Episodes []Episode `json:"episodes"`
		}{
			Episodes: []Episode{
				Episode{
					Name:   "first",
					Season: int64(1), // fine, exists on disk
				},
				Episode{
					Name:   "second",
					Season: int64(1), // fine, exists on disk
				},
				Episode{
					Name:   "third",
					Season: int64(1), // not fine, absent on disk
				},
				Episode{
					Name:   "first in second",
					Season: int64(2), // not fine, absent on disk
				},
			},
		},
	}

	writeShowJSON(TvMazeShow)
}

func TestCreateSeasonJSON(t *testing.T) {
	// step 3; having click on A season
	// name;  season.json
	// url; https://foo.bar/foo/1
	copyR("testdata/Videos_template", "testdata/Videos")
	defer os.RemoveAll("testdata/Videos")

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
}

func TestCreateEpisodeJSON(t *testing.T) {
	// step 4; having click on A episode
	// name; episode.json
	// url; https://foo.bar/foo/1/pilot
	copyR("testdata/Videos_template", "testdata/Videos")
	defer os.RemoveAll("testdata/Videos")

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
}

//func TestMapFoundShowToDiskContent(t *testing.T) {
//file, err := os.Open("testdata/searchResults.json")
//require.NoError(t, err)

//show := &TvMazeShow{}
//require.NoError(t, json.NewDecoder(file).Decode(show))

//fmt.Printf("show = %+v\n", show)
//}

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
