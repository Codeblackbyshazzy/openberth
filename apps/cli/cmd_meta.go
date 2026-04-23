package main

import (
	"fmt"
	"os"
	"runtime"
)

func cmdVersion() {
	fmt.Printf("  %s⚓ OpenBerth%s\n", cBold, cReset)
	fmt.Printf("  CLI version:  %s\n", version)
	fmt.Printf("  Platform:     %s/%s\n", runtime.GOOS, runtime.GOARCH)

	// Try to get server version
	client, err := NewAPIClient()
	if err != nil {
		fmt.Printf("  Server:       %s(not configured)%s\n", cDim, cReset)
		fmt.Println()
		return
	}
	result, err := client.Request("GET", "/health")
	if err != nil {
		fmt.Printf("  Server:       %s(unreachable)%s\n", cDim, cReset)
		fmt.Println()
		return
	}
	if sv, ok := result["version"].(string); ok {
		fmt.Printf("  Server:       %s\n", sv)
	}
	if domain, ok := result["domain"].(string); ok {
		fmt.Printf("  Domain:       %s\n", domain)
	}
	fmt.Println()
}

func cmdConfig() {
	if len(os.Args) < 3 {
		printConfigHelp()
		return
	}

	action := os.Args[2]
	switch action {
	case "set":
		if len(os.Args) < 5 {
			fail("Usage: berth config set <key> <value>")
			os.Exit(1)
		}
		key, value := os.Args[3], os.Args[4]
		cfg := loadCLIConfig()
		switch key {
		case "server":
			cfg.Server = value
		case "key":
			cfg.Key = value
		default:
			fail("Unknown config key: " + key + ". Use 'server' or 'key'.")
			os.Exit(1)
		}
		saveCLIConfig(cfg)
		display := value
		if key == "key" && len(value) > 10 {
			display = value[:10] + "..."
		}
		ok(fmt.Sprintf("Set %s = %s", key, display))

	case "get":
		if len(os.Args) < 4 {
			fail("Usage: berth config get <key>")
			os.Exit(1)
		}
		cfg := loadCLIConfig()
		switch os.Args[3] {
		case "server":
			fmt.Println(cfg.Server)
		case "key":
			fmt.Println(cfg.Key)
		default:
			fmt.Println("(not set)")
		}

	case "show":
		cfg := loadCLIConfig()
		display := cfg.Key
		if len(display) > 10 {
			display = display[:10] + "..."
		}
		fmt.Printf("  server: %s\n", cfg.Server)
		fmt.Printf("  key:    %s\n", display)

	default:
		printConfigHelp()
	}
}

func printConfigHelp() {
	fmt.Println("Usage:")
	fmt.Println("  berth config set <key> <value>")
	fmt.Println("  berth config get <key>")
	fmt.Println("  berth config show")
}
