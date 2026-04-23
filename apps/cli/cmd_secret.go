package main

import (
	"fmt"
	"os"
)

func cmdSecret() {
	subArgs := os.Args[2:]
	if len(subArgs) == 0 {
		fail("Usage: berth secret <set|list|delete>")
		os.Exit(1)
	}

	switch subArgs[0] {
	case "set":
		cmdSecretSet(subArgs[1:])
	case "list", "ls":
		cmdSecretList()
	case "delete", "rm":
		cmdSecretDelete(subArgs[1:])
	default:
		fail("Unknown secret command: " + subArgs[0])
		os.Exit(1)
	}
}

func cmdSecretSet(setArgs []string) {
	if len(setArgs) < 2 {
		fail("Usage: berth secret set NAME VALUE [--description \"desc\"] [--global]")
		os.Exit(1)
	}

	name := setArgs[0]
	value := setArgs[1]

	// Parse optional flags from remaining args
	description := ""
	global := false
	for i := 2; i < len(setArgs); i++ {
		switch setArgs[i] {
		case "--description":
			if i+1 < len(setArgs) {
				i++
				description = setArgs[i]
			}
		case "--global":
			global = true
		}
	}

	fmt.Println()
	fmt.Printf("  %s⚓ OpenBerth Secret Set%s\n\n", cBold, cReset)

	client, err := NewAPIClient()
	if err != nil {
		fail(err.Error())
		os.Exit(1)
	}

	body := map[string]interface{}{
		"name":   name,
		"value":  value,
		"global": global,
	}
	if description != "" {
		body["description"] = description
	}

	result, err := client.RequestJSON("POST", "/api/secrets", body)
	if err != nil {
		fail(err.Error())
		os.Exit(1)
	}

	msg, _ := result["message"].(string)
	if msg == "" {
		msg = "Secret saved"
	}
	ok(msg)

	// Show restarted deployments if any
	if restarted, ok2 := result["restarted"].([]interface{}); ok2 && len(restarted) > 0 {
		for _, d := range restarted {
			if name, ok3 := d.(string); ok3 {
				info(fmt.Sprintf("Restarted: %s", name))
			}
		}
	}
	fmt.Println()
}

func cmdSecretList() {
	fmt.Println()
	fmt.Printf("  %s⚓ OpenBerth Secrets%s\n\n", cBold, cReset)

	client, err := NewAPIClient()
	if err != nil {
		fail(err.Error())
		os.Exit(1)
	}

	result, err := client.Request("GET", "/api/secrets")
	if err != nil {
		fail(err.Error())
		os.Exit(1)
	}

	secrets, _ := result["secrets"].([]interface{})
	if len(secrets) == 0 {
		info("No secrets found.")
		fmt.Println()
		return
	}

	// Print table header
	fmt.Printf("  %s%-20s %-10s %-30s %s%s\n", cBold, "NAME", "SCOPE", "DESCRIPTION", "UPDATED", cReset)
	fmt.Printf("  %s%-20s %-10s %-30s %s%s\n", cDim, "----", "-----", "-----------", "-------", cReset)

	for _, s := range secrets {
		sec, ok2 := s.(map[string]interface{})
		if !ok2 {
			continue
		}
		sName, _ := sec["name"].(string)
		scope, _ := sec["scope"].(string)
		desc, _ := sec["description"].(string)
		updated, _ := sec["updatedAt"].(string)

		fmt.Printf("  %-20s %-10s %-30s %s\n", sName, scope, desc, updated)
	}
	fmt.Println()
}

func cmdSecretDelete(delArgs []string) {
	if len(delArgs) < 1 {
		fail("Usage: berth secret delete NAME [--global]")
		os.Exit(1)
	}

	name := delArgs[0]
	global := false
	for _, a := range delArgs[1:] {
		if a == "--global" {
			global = true
		}
	}

	fmt.Println()
	fmt.Printf("  %s⚓ OpenBerth Secret Delete%s\n\n", cBold, cReset)

	client, err := NewAPIClient()
	if err != nil {
		fail(err.Error())
		os.Exit(1)
	}

	path := "/api/secrets/" + name
	if global {
		path += "?global=true"
	}

	_, err = client.Request("DELETE", path)
	if err != nil {
		fail(err.Error())
		os.Exit(1)
	}

	ok(fmt.Sprintf("Secret %s%s%s deleted", cBold, name, cReset))
	fmt.Println()
}
