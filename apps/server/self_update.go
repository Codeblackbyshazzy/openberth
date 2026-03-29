package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const githubRepo = "AmirSoleimani/openberth"

func selfUpdate(args []string) {
	checkOnly := false
	targetVersion := ""

	for i, a := range args {
		switch a {
		case "--check":
			checkOnly = true
		case "--version":
			if i+1 < len(args) {
				targetVersion = args[i+1]
			}
		}
	}

	fmt.Printf("Current version: %s\n", version)

	if targetVersion == "" {
		// Fetch latest release from GitHub API
		latest, err := fetchLatestVersion()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check for updates: %v\n", err)
			os.Exit(1)
		}
		targetVersion = latest
	}

	fmt.Printf("Latest version:  %s\n", targetVersion)

	// Normalize for comparison (strip 'v' prefix)
	current := strings.TrimPrefix(version, "v")
	target := strings.TrimPrefix(targetVersion, "v")

	if current == target {
		fmt.Println("Already up to date.")
		return
	}

	if checkOnly {
		fmt.Printf("\nUpdate available: %s → %s\n", version, targetVersion)
		fmt.Println("Run 'sudo berth-server update' to install.")
		return
	}

	// Check we can write to our own binary
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine binary path: %v\n", err)
		os.Exit(1)
	}

	// Download new binary
	arch := runtime.GOARCH
	filename := fmt.Sprintf("berth-server-linux-%s", arch)
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", githubRepo, targetVersion, filename)

	fmt.Printf("Downloading %s...\n", filename)

	tmpFile, err := os.CreateTemp("", "berth-server-update-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp file: %v\n", err)
		os.Exit(1)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // cleanup on failure

	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "Download failed: HTTP %d (check version tag exists)\n", resp.StatusCode)
		os.Exit(1)
	}

	n, err := io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Downloaded %.1f MB\n", float64(n)/1024/1024)

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to chmod: %v\n", err)
		os.Exit(1)
	}

	// Replace binary
	if err := os.Rename(tmpPath, exePath); err != nil {
		// Cross-device rename — copy instead
		src, err2 := os.Open(tmpPath)
		if err2 != nil {
			fmt.Fprintf(os.Stderr, "Failed to open downloaded binary: %v\n", err2)
			os.Exit(1)
		}
		dst, err2 := os.OpenFile(exePath, os.O_WRONLY|os.O_TRUNC, 0755)
		if err2 != nil {
			src.Close()
			fmt.Fprintf(os.Stderr, "Failed to write binary (try with sudo): %v\n", err2)
			os.Exit(1)
		}
		io.Copy(dst, src)
		src.Close()
		dst.Close()
	}

	fmt.Printf("Updated to %s\n", targetVersion)

	// Restart systemd service if running under it
	if isSystemdService() {
		fmt.Println("Restarting openberth service...")
		if err := exec.Command("systemctl", "restart", "openberth").Run(); err != nil {
			fmt.Printf("Warning: could not restart service: %v\n", err)
			fmt.Println("Restart manually: sudo systemctl restart openberth")
		} else {
			fmt.Println("Service restarted")
		}
	}
}

func fetchLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	if release.TagName == "" {
		return "", fmt.Errorf("no releases found")
	}

	return release.TagName, nil
}

func isSystemdService() bool {
	// Check if we're running under systemd by looking for INVOCATION_ID env
	// or checking if the openberth service exists
	if os.Getenv("INVOCATION_ID") != "" {
		return true
	}
	err := exec.Command("systemctl", "is-active", "--quiet", "openberth").Run()
	return err == nil
}
