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
	"github.com/gorilla/websocket"
)

const relayServer = "wss://skyping-production.up.railway.app"

func Start() {
	code := generateCode()

	fmt.Println()
	fmt.Println("  Skyping agent running")
	fmt.Printf("  Your code: %s %s\n", code[:3], code[3:])
	fmt.Println()
	fmt.Println("  Share this code with your teammate.")
	fmt.Printf("  They open: https://jeetkumar.space/connect.html\n")
	fmt.Println()
	fmt.Println("  Press Ctrl+C to stop.")
	fmt.Println()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		fmt.Println("\n  Session ended.")
		os.Exit(0)
	}()

	for {
		fmt.Println("  Connecting to relay server...")
		conn, _, err := websocket.DefaultDialer.Dial(relayServer+"/agent/"+code, nil)
		if err != nil {
			fmt.Printf("  Error: %v — retrying in 3s...\n", err)
			time.Sleep(3 * time.Second)
			continue
		}

		fmt.Println("  Connected to relay. Waiting for client...")
		fmt.Println("  Starting terminal session...")
		handleSession(conn)
		conn.Close()
		fmt.Println("  Session ended. Reconnecting...")
		time.Sleep(1 * time.Second)
	}
}

func handleSession(conn *websocket.Conn) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	fmt.Printf("  Starting shell: %s\n", shell)

	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color", "COLUMNS=220", "LINES=50")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		fmt.Printf("  Failed to start shell: %v\n", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Failed to start shell\r\n"))
		return
	}
	defer ptmx.Close()

	pty.Setsize(ptmx, &pty.Winsize{Rows: 50, Cols: 220})
	fmt.Println("  Shell started. Bridging...")

	done := make(chan struct{}, 2)

	// pty → relay
	go func() {
		defer func() { done <- struct{}{} }()
		buf := make([]byte, 1024)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				fmt.Printf("  PTY read error: %v\n", err)
				return
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				fmt.Printf("  WS write error: %v\n", err)
				return
			}
		}
	}()

	// relay → pty
	go func() {
		defer func() { done <- struct{}{} }()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Printf("  WS read error: %v\n", err)
				return
			}
			ptmx.Write(msg)
		}
	}()

	<-done
	cmd.Process.Kill()
	cmd.Wait()
	fmt.Println("  Shell exited.")
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
		fmt.Printf("\n  Could not connect.\n")
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("  Connected!")
	streamTerminal(conn)
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
