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

	"github.com/0xAX/notificator"
)

var notify *notificator.Notificator
var DB = os.Getenv("GOPATH") + "/src/github.com/FenwickElliott/GoSnatch/db/"

func main() {
	accessBearer, err := ioutil.ReadFile(DB + "accessBearer")

	if err != nil {
		initialize()
	} else {
		os.Setenv("AccessBearer", string(accessBearer))
	}

	songID, songName := getSong()
	playlistID := getPlaylist()
	userID := getMe()

	sucsess := goSnatch(userID, songID, playlistID)

	if sucsess {
		notify = notificator.New(notificator.Options{
			DefaultIcon: "icon/default.png",
			AppName:     "GoSnatch",
		})

		msg := songName + " was sucsessfully GoSnatched!!!"

		notify.Push("Horray", msg, "/home/user/icon.png", notificator.UR_CRITICAL)
	}
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

func getMe() string {
	me := get("me")
	return me["id"].(string)
}

func getPlaylist() string {
	list := get("me/playlists")
	items := list["items"].([]interface{})

	for _, v := range items {
		cell := v.(map[string]interface{})
		if cell["name"] == "GoSnatch" {
			return cell["id"].(string)
		}
	}
	// create playlist
	return "Playlist not found"
}

func getSong() (string, string) {
	song := get("me/player/currently-playing")
	item := song["item"].(map[string]interface{})
	return item["id"].(string), item["name"].(string)
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
		panic("\n\nApperetnly no song is playing, sorry\n\nMust fix this more gracefully")
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
	target := DB + name
	f, _ := os.Create(target)
	f.WriteString(content)
	defer f.Close()
}

func refresh() {
	fmt.Println("refreshing")
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
