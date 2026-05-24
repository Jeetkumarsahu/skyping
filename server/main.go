package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
        EnableCompression: false,
}

type Session struct {
	agent  *websocket.Conn
	client chan *websocket.Conn
}

var (
	sessions = make(map[string]*Session)
	mu       sync.Mutex
)

func main() {
	http.HandleFunc("/agent/", handleAgent)
	http.HandleFunc("/client/", handleClient)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	port := "8080"
	fmt.Printf("Skyping relay server running on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleAgent(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[len("/agent/"):]
	if code == "" {
		http.Error(w, "code required", 400)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	session := &Session{
		agent:  conn,
		client: make(chan *websocket.Conn, 1),
	}

	mu.Lock()
	sessions[code] = session
	mu.Unlock()

	fmt.Printf("Agent connected: %s\n", code)

	defer func() {
		mu.Lock()
		delete(sessions, code)
		mu.Unlock()
		fmt.Printf("Agent disconnected: %s\n", code)
	}()

	// Wait for clients — agent stays alive between sessions
	for {
		client := <-session.client
		fmt.Printf("Bridging: %s\n", code)
		bridge(conn, client)
		fmt.Printf("Client disconnected, waiting for new client: %s\n", code)
	}
}

func handleClient(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[len("/client/"):]
	if code == "" {
		http.Error(w, "code required", 400)
		return
	}

	mu.Lock()
	session := sessions[code]
	mu.Unlock()

	if session == nil {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		conn.WriteMessage(websocket.TextMessage, []byte("invalid code"))
		conn.Close()
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	fmt.Printf("Client connected: %s\n", code)

	// Signal agent that client connected
	session.agent.WriteMessage(websocket.TextMessage, []byte("connected"))

	session.client <- conn
}

func bridge(agent, client *websocket.Conn) {
	done := make(chan struct{}, 2)

	// agent → client
	go func() {
		defer func() { done <- struct{}{} }()
		for {
			mt, msg, err := agent.ReadMessage()
			if err != nil {
				return
			}
			if err := client.WriteMessage(mt, msg); err != nil {
				return
			}
		}
	}()

	// client → agent
	go func() {
		defer func() { done <- struct{}{} }()
		for {
			mt, msg, err := client.ReadMessage()
			if err != nil {
				return
			}
			if err := agent.WriteMessage(mt, msg); err != nil {
				return
			}
		}
	}()

	<-done
	client.Close()
	// Agent stays alive for next client
}
