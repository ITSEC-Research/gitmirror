package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/joho/godotenv"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type Repository struct {
    ID     int    `json:"id"`
    Source string `json:"source"`
    Target string `json:"target"`
}

func logMessage(message string) {
    log.Printf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), message)
}

func errorMessage(message string) {
    log.Fatalf("[%s] ERROR: %s", time.Now().Format("2006-01-02 15:04:05"), message)
}

func syncRepositories(source, target, sourceToken, targetToken, username1, username2 string) {
	// Construct authenticated URLs
	sourceURL := fmt.Sprintf("http://%s:%s@%s", username2, sourceToken, source)
	targetURL := fmt.Sprintf("https://%s:%s@%s", username1, targetToken, target)

	logMessage(fmt.Sprintf("Syncing %s to %s", source, target))

	// Clone the source repository
	tmpDir, err := ioutil.TempDir("", "git-sync-")
	if err != nil {
		errorMessage(fmt.Sprintf("Failed to create temporary directory: %s", err))
	}
	defer os.RemoveAll(tmpDir)

	repo, err := git.PlainClone(tmpDir, true, &git.CloneOptions{
		URL: sourceURL,
		Auth: &http.BasicAuth{
			Username: username2,
			Password: sourceToken,
		},
	})
	if err != nil {
		errorMessage(fmt.Sprintf("Failed to clone %s: %s", sourceURL, err))
	}

	// Add the target remote
	remote, err := repo.CreateRemote(&config.RemoteConfig{
		Name: "target",
		URLs: []string{targetURL},
	})
	if err != nil {
		errorMessage(fmt.Sprintf("Failed to add target remote: %s", err))
	}

	// Push all branches to the target repository
	err = remote.Push(&git.PushOptions{
		RemoteName: "target",
		Auth: &http.BasicAuth{
			Username: username1,
			Password: targetToken,
		},
		Force: true,
	})

	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			logMessage("All branches are already up-to-date")
		} else {
			errorMessage(fmt.Sprintf("Failed to push all branches to target: %s", err))
		}
	}

	logMessage("Sync completed")
}
func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		errorMessage("Error loading .env file")
	}
    // Define authentication variables from environment variables
    username1 := os.Getenv("GIT_USERNAME1")
    username2 := os.Getenv("GIT_USERNAME2")
    sourceToken := os.Getenv("GIT_SOURCE_TOKEN")
    targetToken := os.Getenv("GIT_TARGET_TOKEN")

    if username1 == "" || username2 == "" || sourceToken == "" || targetToken == "" {
        errorMessage("One or more required environment variables are missing")
    }

    // Read JSON file
    jsonFile, err := os.Open("repositories.json")
    if err != nil {
        errorMessage(fmt.Sprintf("Failed to open JSON file: %s", err))
    }
    defer jsonFile.Close()

    // Read the file content
    byteValue, err := ioutil.ReadAll(jsonFile)
    if err != nil {
    	errorMessage(fmt.Sprintf("Failed to read JSON file: %s", err))
    }

    // Parse JSON content
    var repositories []Repository
    if err := json.Unmarshal(byteValue, &repositories); err != nil {
        errorMessage(fmt.Sprintf("Failed to parse JSON: %s", err))
    }

    // Process each repository
    for _, repo := range repositories {
        logMessage(fmt.Sprintf("Processing repository ID %d", repo.ID))
        syncRepositories(repo.Source, repo.Target, sourceToken, targetToken, username1, username2)
    }
}
