// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slack

import (
	"encoding/json"
	"fmt"
	"github.com/alcomist/go-portfolio/internal/config"
	"io"
	"log"
	"net/http"
	"strings"
)

type Payload struct {
	Text string `json:"text"`
}

func getConfig() map[string]string {

	cfg := make(map[string]string)

	section := config.MustGet("slack")

	for _, key := range section.Keys() {
		cfg[key.Name()] = section.Key(key.Name()).String()
	}

	return cfg
}

func Post(channel, s string) {

	slackConfig := getConfig()

	url, ok := slackConfig[channel]
	if ok {

		text := Payload{Text: s}
		data, err := json.Marshal(text)
		if err != nil {
			panic(fmt.Errorf("json marshalling error : %v\n", err))
		}

		payload := strings.NewReader(string(data))

		req, _ := http.NewRequest("POST", url, payload)
		req.Header.Add("content-type", "application/json")

		res, _ := http.DefaultClient.Do(req)

		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		if string(body) != "ok" {
			log.Printf("slack send error : %v\n", body)
		}

		return
	}

	log.Println("no channel in slack config")
}
