package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	root := os.Args[1]

	log.Fatal(http.ListenAndServe(":8082", http.FileServer(http.Dir(root))))
}
