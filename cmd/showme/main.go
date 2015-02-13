package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Show struct {
	URL     url.URL  `json:"url"`
	Name    string   `json:"name"`
	Seasons []Season `json:"seasons"`
}

type Season struct {
	Season   int       `json:"season"`
	Episodes []Episode `json:"episodes"`
}

type Episode struct {
	Name    string    `json:"name"`
	Episode int       `json:"episode"`
	AirDate time.Time `json:"airdate"`
}

type Index struct {
	Shows []Show `json:"shows"`
}

func NewIndex() *Index {
	fs, err := ioutil.ReadDir(root)
	if err != nil {
		return nil
	}

	var index = Index{}
	for _, fileinfo := range fs {
		if fileinfo.IsDir() {
			index.Shows = append(index.Shows, Show{Name: fileinfo.Name()})
		}
	}
	return &index
}

func showsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("it works"))
	})
}

//func showHandler() http.Handler {
//return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//})
//}

//func seasonsHandler() http.Handler {
//return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//})
//}

//func seasonHandler() http.Handler {
//return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//})
//}

//func episodesHandler() http.Handler {
//return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//})
//}

//func episodeHandler() http.Handler {
//return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//})
//}

//func moviesHandler() http.Handler {
//return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//})
//}

//func movieHandler() http.Handler {
//return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//})
//}

//func withLogger(l *log.Logger, next http.Handler) http.Handler {
//return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//next.ServeHTTP(w, r)
//})
//}

func main() {
	root := os.Args[1]

	log.Fatal(http.ListenAndServe(":8080", http.FileServer(http.Dir(root))))

}
