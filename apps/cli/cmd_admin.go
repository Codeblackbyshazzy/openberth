package main

import (
	"fmt"
	"os"
	"time"
)

func cmdBackup() {
	fmt.Println()
	fmt.Printf("  %s⚓ OpenBerth Backup%s\n\n", cBold, cReset)

	output := getFlag("output", fmt.Sprintf("openberth-backup-%s.tar.gz", time.Now().Format("2006-01-02")))

	spin("Downloading backup")
	client, err := NewAPIClient()
	if err != nil {
		done()
		fail(err.Error())
		os.Exit(1)
	}

	size, err := client.Download("/api/admin/backup", output)
	if err != nil {
		done()
		fail(err.Error())
		os.Exit(1)
	}
	done()

	ok(fmt.Sprintf("Backup saved: %s%s%s (%s)", cBold, output, cReset, formatSize(size)))
	fmt.Println()
}

func cmdRestore() {
	if len(os.Args) < 3 {
		fail("Usage: berth restore <backup-file.tar.gz>")
		os.Exit(1)
	}
	backupFile := os.Args[2]

	if _, err := os.Stat(backupFile); err != nil {
		fail("File not found: " + backupFile)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("  %s⚓ OpenBerth Restore%s\n\n", cBold, cReset)

	spin("Uploading backup and restoring")
	client, err := NewAPIClient()
	if err != nil {
		done()
		fail(err.Error())
		os.Exit(1)
	}

	result, err := client.UploadFile("/api/admin/restore", backupFile, "backup")
	if err != nil {
		done()
		fail(err.Error())
		os.Exit(1)
	}
	done()

	msg, _ := result["message"].(string)
	users, _ := result["users"].(float64)
	deploys, _ := result["deployments"].(float64)
	rebuilding, _ := result["rebuilding"].(float64)

	ok("Backup restored successfully.")
	info(fmt.Sprintf("Users: %d", int(users)))
	info(fmt.Sprintf("Deployments: %d", int(deploys)))
	if rebuilding > 0 {
		info(fmt.Sprintf("Rebuilding: %d deployment(s) in background", int(rebuilding)))
		warn("TLS certificates may take a few minutes to provision — expect brief SSL errors until then.")
	}
	_ = msg
	fmt.Println()
}
