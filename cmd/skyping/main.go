package main

import (
	"fmt"
	"os"

	"jeetkumar.space/skyping/internal/agent"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "agent":
		agent.Start()
	case "connect":
		if len(os.Args) < 3 {
			fmt.Println("Usage: skyping connect <6-digit-code>")
			os.Exit(1)
		}
		agent.Connect(os.Args[2])
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
  skyping agent              Start agent, get your 6-digit code
  skyping connect <code>     Connect to a remote terminal session
  skyping --version          Show version
  skyping --help             Show this help
`)
}
