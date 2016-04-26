package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
)

var tvMazeURLTemplate = "http://api.tvmaze.com/singlesearch/shows?q=%s&embed=episodes"

type TvMazeShow struct {
	Name  string `json:"name"`
	Image struct {
		Medium   string `json:"medium"`
		Original string `json:"original"`
	} `json:"image"`
	Summary  string `json:"summary"`
	Embedded struct {
		Episodes []struct {
			ID      int64  `json:"id"`
			Name    string `json:"name"`
			Season  int64  `json:"season"`
			Episode int64  `json:"episode"`
			Summary string `json:"summary"`
			//AirDate time.Time `json:"airdate"`
			Image struct {
				Medium   string `json:"medium"`
				Original string `json:"original"`
			} `json:"image"`
		} `json:"episodes"`
	} `json:"_embedded"`
}

type TvMazeClient struct {
	URLTemplate string
	logger      *logrus.Entry
}

func (t TvMazeClient) Find(q string) (*TvMazeShow, error) {
	response, err := http.Get(fmt.Sprintf(
		t.URLTemplate, q))
	if err != nil {
		t.logger.WithField("err", err).Error("Failed to get a response")
		return nil, err
	}
	if response.StatusCode == 404 {
		t.logger.Warn("No match found")
		return nil, err
	}

	show := &TvMazeShow{}
	if err := json.NewDecoder(response.Body).Decode(show); err != nil {
		t.logger.WithField("err", err).Error("Failed to decode")
		return nil, err
	}

	return show, nil
}
