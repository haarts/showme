package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	trakt "github.com/42minutes/go-trakt"
	"github.com/stretchr/testify/assert"
)

type TestAuthMethod struct{}

func (t TestAuthMethod) String() string {
	return "god"
}

type mockFileInfo struct{}

func (m mockFileInfo) Name() string       { return "foo" }
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
