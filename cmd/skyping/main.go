package main

import (
	"fmt"
	"os"

	"jeetkumar.space/skyping/internal/agent"
)

const version = "0.4.0"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "agent":
		agent.Start()
	case "--version", "version":
		fmt.Printf("skyping v%s\n", version)
	case "--help", "help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`
Skyping — peer-to-peer terminal sharing

Usage:
  skyping agent              Start agent, share the generated URL
  skyping --version          Show version
  skyping --help             Show this help
`)
}
