package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	fmt.Println("Listening on localhost:8081")
	fmt.Println("'/' serves files located in '../static'")
	fmt.Println("'/shows/' serves files located in '../cmd/fetcher/testdata/example_result/shows'")
	fmt.Println("'/login' and '/register' proxy to authme (which should run on localhost:8080)")

	http.Handle("/", http.FileServer(http.Dir("../static")))
	http.Handle("/shows/", http.FileServer(http.Dir("../cmd/fetcher/testdata/example_result/")))

	rpURL, err := url.Parse("http://localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/login", httputil.NewSingleHostReverseProxy(rpURL))
	http.Handle("/register", httputil.NewSingleHostReverseProxy(rpURL))

	http.ListenAndServe(":8081", nil)
}
