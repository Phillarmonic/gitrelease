package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Version of the application
const Version = "1.1.0"

// GitRelease: Fetches the latest release version of a GitHub repository

// Release represents the latest release information from GitHub
type Release struct {
	TagName string `json:"tag_name"`
}

func main() {
	versionFlag := flag.Bool("version", false, "Print the version of the application")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("GitRelease version %s\n", Version)
		return
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: gitrelease <github repo ex: phillarmonic/gitrelease>")
		fmt.Println("       gitrelease --version")
		os.Exit(0)
	}

	repo := args[0]
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
