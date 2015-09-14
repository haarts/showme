package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

const (
	port = 8082
	host = ""
)

var addr = host + ":" + strconv.Itoa(port)

func main() {
	root, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Error constucting absolute path")
		return
	}

	log.WithFields(log.Fields{
		"root": root,
		"port": port,
		"host": host,
	}).Infof("Serving %s on %s", root, addr)

	log.Fatal(http.ListenAndServe(addr, http.FileServer(http.Dir(root))))
}
