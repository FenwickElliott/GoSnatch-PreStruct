package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/deckarep/gosx-notifier"
)

var db = os.Getenv("GOPATH") + "/src/github.com/FenwickElliott/GoSnatch/db/"

// var db = ""

func main() {
	accessBearer, err := ioutil.ReadFile(db + "accessBearer")

	if err != nil {
		initialize()
	} else {
		os.Setenv("AccessBearer", string(accessBearer))
	}

	cPlaylistID := make(chan string)
	cSong := make(chan []string)
	cUserID := make(chan string)

	go getPlaylist(cPlaylistID)
	go getSong(cSong)
	go getMe(cUserID)

	userID := <-cUserID
	song := <-cSong
	playlistID := <-cPlaylistID

	if checkPlaylist(userID, song[0], playlistID) {
		if goSnatch(userID, song[0], playlistID) {
			sendNote(song[1], song[2], "Was sucsessfully Snatched")
		}
	} else {
		sendNote(song[1], song[2], "Had already been Snatched")
	}
}

func sendNote(title, subtitle, message string) {
	note := gosxnotifier.NewNotification(message)
	note.Title = title
	note.Subtitle = subtitle
	// note.ContentImage = db + "icon.png"
	note.AppIcon = db + "icon.png"
	_ = note.Push()
}

func goSnatch(userID, songID, playlistID string) bool {
	url := "https://api.spotify.com/v1/users/" + userID + "/playlists/" + playlistID + "/tracks?uris=spotify%3Atrack%3A" + songID

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", os.Getenv("AccessBearer"))
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return true
}

func getMe(cUserID chan string) {
	me := get("me")
	cUserID <- me["id"].(string)
	// return me["id"].(string)
}

func getPlaylist(cPlaylistID chan string) {
	list := get("me/playlists")
	items := list["items"].([]interface{})

	for _, v := range items {
		cell := v.(map[string]interface{})
		if cell["name"] == "GoSnatch" {
			// return cell["id"].(string)
			cPlaylistID <- cell["id"].(string)
			return
		}
	}
	cPlaylistID <- createPlaylist()
}

func createPlaylist() string {
	url := "https://api.spotify.com/v1/users/cjgfe/playlists"
	body := strings.NewReader(`{"name":"GoSnatch","description":"Your automatically generated GoSnatch playlist!","public":"false"}`)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Authorization", os.Getenv("AccessBearer"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	map2b := make(map[string]interface{})
	err = json.Unmarshal(bodyBytes, &map2b)

	if err != nil {
		fmt.Println(err)
	}

	return map2b["id"].(string)
}

func checkPlaylist(userID, songID, playlistID string) bool {
	url := "users/" + userID + "/playlists/" + playlistID + "/tracks"
	playlist := get(url)
	tracks := playlist["items"].([]interface{})

	for _, v := range tracks {
		track := v.(map[string]interface{})
		track2 := track["track"].(map[string]interface{})
		if track2["id"] == songID {
			return false
		}
	}
	return true
}

func getSong(cSong chan []string) {
	song := get("me/player/currently-playing")
	if len(song) == 0 {
		fmt.Println("nothing here")
		sendNote("Nothing Here", "Sorry", "...")
		os.Exit(0)
	}
	item := song["item"].(map[string]interface{})
	artists := item["artists"].([]interface{})
	artist := artists[0].(map[string]interface{})
	cSong <- []string{item["id"].(string), item["name"].(string), artist["name"].(string)}
}

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

	if resp.StatusCode == 204 {
		// panic("\n\nApperetnly no song is playing, sorry\n\nMust fix this more gracefully")
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	map2b := make(map[string]interface{})
	err = json.Unmarshal(bodyBytes, &map2b)
	if err != nil {
		// panic(err)
	}
	return map2b
}
