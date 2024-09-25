package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// GitRelease: Fetches the latest release version of a GitHub repository

// Release represents the latest release information from GitHub
type Release struct {
	TagName string `json:"tag_name"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gitrelease <github repo ex: phillarmonic/gitrelease>")
		os.Exit(0)
	}

	repo := os.Args[1]
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to fetch release: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to fetch release: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	fmt.Println(release.TagName)
}
