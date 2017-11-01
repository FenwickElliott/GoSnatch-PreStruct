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
)

func main() {
	accessBearer, err := ioutil.ReadFile("accessBearer")

	if err != nil {
		initialize()
	} else {
		os.Setenv("AccessBearer", string(accessBearer))
	}

	songID, songName := getSong()
	fmt.Println(songID)
	fmt.Println(songName)
}

func getSong() (string, string) {
	song := getReturn("me/player/currently-playing")
	item := song["item"].(map[string]interface{})
	return item["id"].(string), item["name"].(string)
}

func getReturn(endpoint string) map[string]interface{} {
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
		get(endpoint)
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	map2b := make(map[string]interface{})
	err = json.Unmarshal(bodyBytes, &map2b)
	if err != nil {
		panic(err)
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
	f, _ := os.Create(name)
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
