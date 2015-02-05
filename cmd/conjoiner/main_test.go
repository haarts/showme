package main

import (
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

func TestShowIdex(t *testing.T) {
	copyR("testdata/Videos_template", "testdata/Videos")
	defer os.RemoveAll("testdata/Videos")

	shows := map[os.FileInfo]FullShow{
		mockFileInfo{name: "testdata/Videos/show_one"}: FullShow{show: trakt.Show{Title: "Show One"}},
		mockFileInfo{name: "testdata/Videos/show_two"}: FullShow{show: trakt.Show{Title: "Show two"}},
	}

	fmt.Printf("shows %+v\n", shows)
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
