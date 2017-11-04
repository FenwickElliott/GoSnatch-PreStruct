package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

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
	// create playlist
	// return "Playlist not found"
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

func initialize() {
	fmt.Println("Initializing...")
	go serve()
	askAuth()
	time.Sleep(15 * time.Second)
}

func serve() {
	http.HandleFunc("/catch", catch)
	http.ListenAndServe(":3456", nil)
}

func catch(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Thank you, GoSnatch can now access your spotify account.\nYou may close this window.\n")
	code := r.URL.Query()["code"][0]

	exchangeCode(code)
}

func askAuth() {
	url := "https://accounts.spotify.com/authorize/?client_id=715c15fc7503401fb136d6a79079b50c&response_type=code&redirect_uri=http://localhost:3456/catch&scope=user-read-playback-state%20playlist-read-private%20playlist-modify-private"
	exec.Command("open", url).Start()
}

func exchangeCode(code string) {
	temp := "grant_type=authorization_code&code=" + code + "&redirect_uri=http://localhost:3456/catch"
	body := strings.NewReader(temp)
	req, _ := http.NewRequest("POST", "https://accounts.spotify.com/api/token", body)
	req.Header.Set("Authorization", "Basic NzE1YzE1ZmM3NTAzNDAxZmIxMzZkNmE3OTA3OWI1MGM6ZTkxZWZkZDAzNDVkNDlkNTllOGE2ZDc1YjUzZTE2YTE=")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	bodyMap := make(map[string]interface{})

	err2 := json.Unmarshal([]byte(bodyString), &bodyMap)

	if err != nil {
		panic(err2)
	}

	write("accessBearer", "Bearer "+bodyMap["access_token"].(string))
	os.Setenv("AccessBearer", "Bearer "+bodyMap["access_token"].(string))
	write("refreshBody", "grant_type=refresh_token&refresh_token="+bodyMap["refresh_token"].(string))
}

func write(name, content string) {
	target := db + name
	f, _ := os.Create(target)
	f.WriteString(content)
	defer f.Close()
}

func refresh() {
	fmt.Println("Refreshing...")
	refreshBody, err := ioutil.ReadFile("refreshBody")
	if err != nil {
		initialize()
	}
	body := strings.NewReader(string(refreshBody))
	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Basic NzE1YzE1ZmM3NTAzNDAxZmIxMzZkNmE3OTA3OWI1MGM6ZTkxZWZkZDAzNDVkNDlkNTllOGE2ZDc1YjUzZTE2YTE=")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err2 := http.DefaultClient.Do(req)
	if err2 != nil {
		panic(err2)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	bodyMap := make(map[string]interface{})

	err3 := json.Unmarshal(bodyBytes, &bodyMap)
	if err3 != nil {
		panic(err3)
	}

	os.Setenv("AccessBearer", "Bearer "+bodyMap["access_token"].(string))
	write("accessBearer", "Bearer "+bodyMap["access_token"].(string))
}
