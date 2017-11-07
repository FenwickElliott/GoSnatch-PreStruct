package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	gosxnotifier "github.com/deckarep/gosx-notifier"
)

func get(endpoint string) map[string]interface{} {
	url := "https://api.spotify.com/v1/" + endpoint
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", os.Getenv("AccessBearer"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		refresh()
		main()
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	map2b := make(map[string]interface{})
	err = json.Unmarshal(bodyBytes, &map2b)
	if err != nil {
		// panic(err)
	}
	return map2b
}

func sendNote(title, subtitle, message, icon string) {
	note := gosxnotifier.NewNotification(message)
	note.Title = title
	note.Subtitle = subtitle
	// note.ContentImage = db + "spotify.png"
	note.AppIcon = db + icon + ".png"
	_ = note.Push()
}

func write(name, content string) {
	target := db + name
	f, _ := os.Create(target)
	f.WriteString(content)
	defer f.Close()
}
