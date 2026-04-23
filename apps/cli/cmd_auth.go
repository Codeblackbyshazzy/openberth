package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type loginResult struct {
	apiKey string
	name   string
	err    error
}

func cmdLogin() {
	fmt.Println()
	fmt.Printf("  %s⚓ OpenBerth Login%s\n\n", cBold, cReset)

	cfg := loadCLIConfig()
	if cfg.Server == "" {
		fail("Server not configured. Run: berth config set server https://your-domain.com")
		os.Exit(1)
	}
	server := strings.TrimRight(cfg.Server, "/")

	// Start local callback server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fail("Failed to start local server: " + err.Error())
		os.Exit(1)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	callbackURL := fmt.Sprintf("http://localhost:%d/callback", port)

	resultCh := make(chan loginResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><body><h2>Error</h2><p>No login code received.</p></body></html>`)
			resultCh <- loginResult{err: fmt.Errorf("no login code received")}
			return
		}

		// Exchange code for API key
		body := fmt.Sprintf(`{"code":"%s"}`, code)
		resp, err := http.Post(server+"/api/login/exchange", "application/json", strings.NewReader(body))
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><body><h2>Error</h2><p>Failed to exchange login code.</p></body></html>`)
			resultCh <- loginResult{err: fmt.Errorf("exchange failed: %w", err)}
			return
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]string
		if err := json.Unmarshal(respBody, &result); err != nil || result["apiKey"] == "" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><body><h2>Error</h2><p>Invalid response from server.</p></body></html>`)
			resultCh <- loginResult{err: fmt.Errorf("invalid exchange response")}
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body style="font-family:sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;"><div style="text-align:center;"><h2>Login successful!</h2><p>You can close this tab.</p></div></body></html>`)
		resultCh <- loginResult{apiKey: result["apiKey"], name: result["name"]}
	})

	srv := &http.Server{Handler: mux}
	go srv.Serve(listener)

	// Open browser
	loginURL := fmt.Sprintf("%s/login?callback=%s", server, callbackURL)
	spin("Opening browser")
	if err := openBrowser(loginURL); err != nil {
		done()
		info("Could not open browser. Please visit:")
		fmt.Printf("  %s%s%s\n\n", cCyan, loginURL, cReset)
	} else {
		done()
		info(fmt.Sprintf("If browser didn't open: %s%s%s", cCyan, loginURL, cReset))
	}

	spin("Waiting for login")

	// Wait for result with timeout
	select {
	case result := <-resultCh:
		done()
		srv.Shutdown(context.Background())

		if result.err != nil {
			fail(result.err.Error())
			os.Exit(1)
		}

		// Save API key
		cfg.Key = result.apiKey
		saveCLIConfig(cfg)

		fmt.Println()
		ok(fmt.Sprintf("Logged in as %s%s%s", cBold, result.name, cReset))
		fmt.Println()

	case <-time.After(5 * time.Minute):
		done()
		srv.Shutdown(context.Background())
		fail("Login timed out (5 minutes). Please try again.")
		os.Exit(1)
	}
}

func cmdRotateKey() {
	fmt.Println()
	fmt.Printf("  %s⚓ OpenBerth Rotate Key%s\n\n", cBold, cReset)

	client, err := NewAPIClient()
	if err != nil {
		fail(err.Error())
		os.Exit(1)
	}

	spin("Rotating API key")
	result, err := client.RequestJSON("POST", "/api/me/rotate-key", map[string]string{})
	done()
	if err != nil {
		fail(err.Error())
		os.Exit(1)
	}

	newKey, _ := result["apiKey"].(string)
	if newKey == "" {
		fail("Server did not return a new API key.")
		os.Exit(1)
	}

	cfg := loadCLIConfig()
	cfg.Key = newKey
	if err := saveCLIConfig(cfg); err != nil {
		fail("Rotated server-side, but failed to update ~/.berth.json: " + err.Error())
		info(fmt.Sprintf("New key: %s%s%s", cBold, newKey, cReset))
		os.Exit(1)
	}

	ok("API key rotated. Old key is now invalid.")
	info(fmt.Sprintf("New key: %s%s%s", cBold, newKey, cReset))
	warn("Any other machines using the old key must be updated (`berth config set key <new>` or `berth login`).")
	fmt.Println()
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}
