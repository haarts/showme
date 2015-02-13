package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	root := os.Args[1]

	log.Fatal(http.ListenAndServe(":8080", http.FileServer(http.Dir(root))))

}
