package agent

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/creack/pty"
)

func Start() {
	code := generateCode()

	fmt.Println()
	fmt.Println("  Skyping agent running")
	fmt.Printf("  Your code: %s %s\n", code[:3], code[3:])
	fmt.Println("  Waiting for connection...")
	fmt.Println()
	fmt.Println("  Share this code with your teammate.")
	fmt.Println("  Press Ctrl+C to stop.")
	fmt.Println()

	port := codeToPort(code)
	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		fmt.Printf("Error starting listener: %v\n", err)
		os.Exit(1)
	}
	defer ln.Close()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		fmt.Println("\n  Session ended.")
		os.Exit(0)
	}()

	conn, err := ln.Accept()
	if err != nil {
		fmt.Printf("Connection error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("  Connected! Starting terminal session...")
	fmt.Println()

	handleSession(conn)
}

func Connect(code string) {
	if len(code) != 6 {
		fmt.Println("Code must be exactly 6 digits.")
		os.Exit(1)
	}

	fmt.Printf("\n  Connecting with code: %s %s...\n\n", code[:3], code[3:])

	port := codeToPort(code)

	var conn net.Conn
	var err error
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		conn, err = net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 2*time.Second)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
		fmt.Print(".")
	}
	if err != nil {
		fmt.Printf("\n  Could not connect. Make sure the agent is running.\n")
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("  Connected!")
	streamTerminal(conn)
}

func handleSession(conn net.Conn) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		fmt.Printf("Failed to start shell: %v\n", err)
		return
	}
	defer ptmx.Close()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			ptmx.Write(buf[:n])
		}
	}()

	buf := make([]byte, 1024)
	for {
		n, err := ptmx.Read(buf)
		if err != nil {
			break
		}
		conn.Write(buf[:n])
	}

	cmd.Wait()
	fmt.Println("  Session closed.")
}

func streamTerminal(conn net.Conn) {
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			os.Stdout.Write(buf[:n])
		}
	}()

	buf := make([]byte, 1024)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}
		conn.Write(buf[:n])
	}
}

func generateCode() string {
	code := ""
	for i := 0; i < 6; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code += fmt.Sprintf("%d", n)
	}
	return code
}

func codeToPort(code string) int {
	var n int
	fmt.Sscanf(code, "%d", &n)
	return 10000 + (n % 10000)
}
