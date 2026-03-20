package main

import (
	"fmt"
	"os"

	qrcode "github.com/skip2/go-qrcode"
)

// isTerminal returns true if stdout is a terminal (not piped).
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// printQR generates and prints a QR code to the terminal using Unicode half-blocks.
// Silently does nothing if generation fails.
func printQR(text string) {
	qr, err := qrcode.New(text, qrcode.Low)
	if err != nil {
		return
	}
	fmt.Println()
	fmt.Print(qr.ToSmallString(false))
}
