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
}

type Session struct {
	agent  *websocket.Conn
	client *websocket.Conn
	mu     sync.Mutex
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

	mu.Lock()
	sessions[code] = &Session{agent: conn}
	mu.Unlock()

	fmt.Printf("Agent connected: %s\n", code)

	defer func() {
		mu.Lock()
		delete(sessions, code)
		mu.Unlock()
		conn.Close()
		fmt.Printf("Agent disconnected: %s\n", code)
	}()

	// Wait for client then bridge
	for {
		mu.Lock()
		session := sessions[code]
		mu.Unlock()

		if session != nil && session.client != nil {
			bridge(session)
			return
		}

		_, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
	}
}

func handleClient(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[len("/client/"):]
	if code == "" {
		http.Error(w, "code required", 400)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	mu.Lock()
	session := sessions[code]
	if session == nil {
		mu.Unlock()
		conn.WriteMessage(websocket.TextMessage, []byte("invalid code"))
		conn.Close()
		return
	}
	session.client = conn
	mu.Unlock()

	fmt.Printf("Client connected: %s\n", code)
	bridge(session)
}

func bridge(s *Session) {
	done := make(chan struct{})

	// agent → client
	go func() {
		defer close(done)
		for {
			mt, msg, err := s.agent.ReadMessage()
			if err != nil {
				return
			}
			if err := s.client.WriteMessage(mt, msg); err != nil {
				return
			}
		}
	}()

	// client → agent
	go func() {
		for {
			mt, msg, err := s.client.ReadMessage()
			if err != nil {
				return
			}
			if err := s.agent.WriteMessage(mt, msg); err != nil {
				return
			}
		}
	}()

	<-done
}
