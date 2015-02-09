package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	trakt "github.com/42minutes/go-trakt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestAuthMethod struct{}

func (t TestAuthMethod) String() string {
	return "god"
}

type mockFileInfo struct {
	name string
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return 1 }
func (m mockFileInfo) Mode() os.FileMode  { return os.ModeDir }
func (m mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m mockFileInfo) IsDir() bool        { return true }
func (m mockFileInfo) Sys() interface{}   { return nil }

func readTestData(file string) string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("err %+v\n", err)
		os.Exit(1)
	}
	return string(data)
}

func mockClient(URL string) Trakt {
	client := Trakt{
		trakt.NewClientWith(
			URL,
			"",               // userAgent
			"",               // apiKey
			TestAuthMethod{}, // authMethod
			nil,              // http.Client
		),
	}

	return client
}

func TestDirsToShows(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, readTestData("testdata/searchResults.json"))
	}))
	defer ts.Close()

	client := mockClient(ts.URL)

	shows := client.turnDirsIntoShows([]os.FileInfo{mockFileInfo{}})
	assert.NotEmpty(t, shows)
}

func TestShowResultsToShows(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, readTestData("testdata/showDetails.json"))
	}))
	defer ts.Close()

	client := mockClient(ts.URL)

	r := trakt.ShowResult{}
	shows := client.turnShowResultsIntoShows(
		map[os.FileInfo]trakt.ShowResult{
			mockFileInfo{}: r,
		},
	)
	assert.NotEmpty(t, shows)
}

func TestAnnotateShows(t *testing.T) {
	stack := []string{
		"testdata/seasons.json",
		"testdata/episodes_1.json",
	}
	var f string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		f, stack = stack[0], stack[1:len(stack)] // *POP*
		fmt.Fprintln(w, readTestData(f))
	}))
	defer ts.Close()

	client := mockClient(ts.URL)

	shows := map[os.FileInfo]FullShow{
		mockFileInfo{}: FullShow{},
	}
	client.addSeasonsAndEpisodesToShows(shows)
	show := shows[mockFileInfo{}]
	assert.Len(t, show.seasons, 1)
	assert.Len(t, show.seasons[0].episodes, 10)
}

func TestShowFiles(t *testing.T) {
	copyR("testdata/Videos_template", "testdata/Videos")
	defer os.RemoveAll("testdata/Videos")
	err := createJSONs()
	require.NoError(t, err)

	createdJSONs := []string{
		"testdata/Videos/shows.json",
		"testdata/Videos/Show One.json",
		"testdata/Videos/show1/seasons.json",
		"testdata/Videos/show1/1.json",
		"testdata/Videos/show1/2.json",
		"testdata/Videos/show1/3.json",
		"testdata/Videos/show1/1/episodes.json",
		"testdata/Videos/show1/1/Episode 1.json",
		"testdata/Videos/show1/1/Episode 2.json",
		"testdata/Videos/Show Two.json",
	}

	for _, j := range createdJSONs {
		_, err := os.Stat(j)
		assert.Nil(t, err, "%s should exist", j)
	}
}

func TestShowURLs(t *testing.T) {
	copyR("testdata/Videos_template", "testdata/Videos")
	defer os.RemoveAll("testdata/Videos")
	err := createJSONs()
	require.NoError(t, err)

	data, err := ioutil.ReadFile("testdata/Videos/shows.json")
	require.NoError(t, err)

	var shows []FullShow
	err = json.Unmarshal(data, &shows)
	require.NoError(t, err)

	assert.Equal(t, "testdata/Videos/Show One.json", shows[0].URL)
	assert.Equal(t, "testdata/Videos/Show Two.json", shows[1].URL)
}

func TestSeasonURLs(t *testing.T) {
	copyR("testdata/Videos_template", "testdata/Videos")
	defer os.RemoveAll("testdata/Videos")
	err := createJSONs()
	require.NoError(t, err)

	data, err := ioutil.ReadFile("testdata/Videos/show1/seasons.json")
	require.NoError(t, err)

	var seasons []season
	err = json.Unmarshal(data, &seasons)
	require.NoError(t, err)

	assert.Equal(t, "testdata/Videos/show1/1.json", seasons[0].URL)
	assert.Equal(t, "testdata/Videos/show1/2.json", seasons[1].URL)
}

func createJSONs() error {
	shows := map[os.FileInfo]FullShow{
		mockFileInfo{name: "testdata/Videos/show1"}: FullShow{
			show: trakt.Show{Title: "Show One"},
			seasons: []season{
				{
					// You need to drop the package name to address the embedded field.
					Season: trakt.Season{Number: 1},
					episodes: []trakt.Episode{
						{
							Title: "Episode 1",
						},
						{
							Title: "Episode 2",
						},
					},
				},
				{
					Season: trakt.Season{Number: 2},
				},
				{
					Season: trakt.Season{Number: 3},
				},
			},
		},
		mockFileInfo{name: "testdata/Videos/show2"}: FullShow{
			show: trakt.Show{Title: "Show Two"},
		},
	}

	c := newConjoiner("testdata/Videos")
	return c.createJSONs(shows)
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
