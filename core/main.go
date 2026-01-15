// Main entry point for the AppRun platform CLI.
// This file only contains the main() function which delegates to the cmd package.
package main

import (
	"log"
	"os"

	"apprun/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
