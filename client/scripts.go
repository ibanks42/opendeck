package main

import (
	"encoding/json"
	"io"
	"net/http"
)

func getScripts(hostname, port string) ([]string, error) {
	response, err := http.Get("http://" + hostname + ":" + port + "/scripts")
	if err != nil {
		setFallbackContainer(0, "Failed to load tasks. Try again?")
		return []string{}, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return []string{}, err
	}

	var scripts []string
	if err := json.Unmarshal(body, &scripts); err != nil {
		return []string{}, err
	}
	return scripts, nil
}
