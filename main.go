package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// Build-time variables
var BOT_TOKEN string
var USERNAME string

var cmd = exec.Command("ngrok", "tcp", "22")

type ApiResponse struct {
	StatusCode string `json:"status_code"`
	Tunnels    []struct {
		PublicUrl string `json:"public_url"`
	}
}

func startNgrok() {
	go cmd.Run()
}

func getNgrokUrl() (string, error) {
	cmd := exec.Command("curl", "localhost:4040/api/tunnels")

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("curling api: %w", err)
	}

	var response ApiResponse
	err = json.Unmarshal(out, &response)
	if err != nil {
		return "", fmt.Errorf("parsing api response: %w", err)
	}

	if len(response.Tunnels) < 1 {
		return "", fmt.Errorf("no tunnels in the response")
	}

	return response.Tunnels[0].PublicUrl, nil
}

func sendTelegramMessage(message string) error {
	res, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s", BOT_TOKEN, USERNAME, message))
	if err != nil {
		return fmt.Errorf("making http get request: %w", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if !strings.Contains(string(body), "\"ok\":true") {
		return fmt.Errorf("seems like status is not 200: %s", string(body))
	}

	return nil
}

func chechUrl(previousUrl string) (string, error) {
	publicUrl, err := getNgrokUrl()
	if err != nil {
		return "", fmt.Errorf("getting ngrok public url: %w", err)
	}

	if publicUrl != previousUrl {
		var message string
		if previousUrl != "" {
			message = fmt.Sprintf("Url changed: %s", publicUrl)
		} else {
			message = fmt.Sprintf("New url: %s", publicUrl)
		}

		err := sendTelegramMessage(message)
		if err != nil {
			return "", fmt.Errorf("sending new url via telegram: %w", err)
		}
	}

	return publicUrl, nil
}

func monitNgrok() {
	previousUrl := ""
	retryTimeout := time.Second * 10

    for range time.Tick(time.Second * 10) {
		url, err := chechUrl(previousUrl)

		// Restart everything if failed
		if err != nil {
			errorMessage := fmt.Errorf("checking ngrok tunnel url: %w", err)
			fmt.Println(errorMessage)
			err := sendTelegramMessage(errorMessage.Error())
			if err != nil {
				fmt.Println("could not send error message :c", err)
			}

			time.Sleep(retryTimeout)
			if (retryTimeout < time.Minute * 15) {
				retryTimeout += time.Minute * 1
			}

			cmd := exec.Command("killall", "ngrok")
			err = cmd.Run()
			if err != nil {
				fmt.Println("killing ngrok:", err)
				err := sendTelegramMessage(err.Error())
				if err != nil {
					fmt.Println("could not send error message :c", err)
				}
			}

			startNgrok()

			monitNgrok()
		}

		retryTimeout = time.Second * 10
		previousUrl = url
    }
}


func main() {
	fmt.Println("Starting ngrok")
	startNgrok()

	time.Sleep(time.Second * 2)
	fmt.Println("Starting to monitor ngrok")

	monitNgrok()
}
