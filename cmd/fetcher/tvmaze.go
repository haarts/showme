package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
)

type TvMazeEpisode struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Season  int64  `json:"season"`
	Episode int64  `json:"number"`
	Summary string `json:"summary"`
	//AirDate time.Time `json:"airdate"`
	Image struct {
		Medium   string `json:"medium"`
		Original string `json:"original"`
	} `json:"image"`
}

type TvMazeShow struct {
	Name  string `json:"name"`
	Image struct {
		Medium   string `json:"medium"`
		Original string `json:"original"`
	} `json:"image"`
	Summary  string `json:"summary"`
	Embedded struct {
		Episodes []TvMazeEpisode `json:"episodes"`
	} `json:"_embedded"`
}

type TvMazeClient struct {
	logger *logrus.Entry
}

func (t TvMazeClient) Find(q string) (*TvMazeShow, error) {
	query := fmt.Sprintf(os.Getenv("TVMAZE_URL_TEMPLATE"), q)
	contextLogger := t.logger.WithField("url", query)
	contextLogger.Debug("Querying TVMaze")

	response, err := http.Get(query)
	if err != nil {
		contextLogger.WithField("err", err).Error("Failed to get a response")
		return nil, err
	}
	if response.StatusCode == 404 {
		contextLogger.Warn("No match found")
		return nil, err
	}

	show := &TvMazeShow{}
	if err := json.NewDecoder(response.Body).Decode(show); err != nil {
		contextLogger.WithField("err", err).Error("Failed to decode")
		return nil, err
	}

	return show, nil
}
