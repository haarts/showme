package main

import (
  "net/http"
  "net/http/httputil"
  "net/url"
  "fmt"
  "log"
)

func main() {
  fmt.Println("This server serves files located in '../static'")
	http.Handle("/", http.FileServer(http.Dir("../static")))

	rpURL, err := url.Parse("http://localhost:8080")
	if err != nil {
  	    log.Fatal(err)
	}
	http.Handle("/login", httputil.NewSingleHostReverseProxy(rpURL))

	http.ListenAndServe(":8081", nil)
}
