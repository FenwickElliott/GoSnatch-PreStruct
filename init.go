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
)

func initialize() {
	fmt.Println("Initializing...")
	done := make(chan bool)
	go serve(done)
	askAuth()

	finished := <-done

	if finished {
		fmt.Println("Initiation complete")
	}
}

func serve(done chan bool) {
	http.HandleFunc("/catch", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Thank you, GoSnatch can now access your spotify account.\nYou may close this window.\n")
		code := r.URL.Query()["code"][0]

		exchangeCode(code, done)
	})
	http.ListenAndServe(":3456", nil)
}

func askAuth() {
	url := "https://accounts.spotify.com/authorize/?client_id=715c15fc7503401fb136d6a79079b50c&response_type=code&redirect_uri=http://localhost:3456/catch&scope=user-read-playback-state%20playlist-read-private%20playlist-modify-private"
	exec.Command("open", url).Start()
}

func exchangeCode(code string, done chan bool) {
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

	done <- true
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
