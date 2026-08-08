package agent

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"jeetkumar.space/skyping/internal/tunnel"
)

const (
	connectPageURL   = "https://jeetkumar.space/connect.html"
	handshakeTimeout = 15 * time.Second
	writeTimeout     = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return r.Method == http.MethodGet
	},
}

type session struct {
	privateKey *ecdh.PrivateKey
	publicKey  []byte
	token      []byte
	active     int32
}

type clientHello struct {
	PublicKey string `json:"publicKey"`
	Proof     string `json:"proof"`
}

type serverHello struct {
	Proof string `json:"proof"`
}

func Start() {
	privateKey, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate session key: %v", err)
	}

	token := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, token); err != nil {
		log.Fatalf("Failed to generate session secret: %v", err)
	}

	currentSession := &session{
		privateKey: privateKey,
		publicKey:  privateKey.PublicKey().Bytes(),
		token:      token,
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("Failed to start local server: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", currentSession.handleWebSocket)
	server := &http.Server{Handler: mux}
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("Local server error: %v", err)
		}
	}()

	fmt.Println("  Starting private tunnel via cloudflared...")
	tunnelURL, tunnelCmd, err := tunnel.StartTunnel(listener.Addr().(*net.TCPAddr).Port)
	if err != nil {
		_ = server.Close()
		<-serverDone
		log.Fatalf("Failed to start private tunnel: %v\n  Install cloudflared and try again.", err)
	}

	shareLink := buildShareLink(tunnelURL, currentSession.publicKey, currentSession.token)
	printShareLink(shareLink)

	cleanupSignal := make(chan os.Signal, 1)
	signal.Notify(cleanupSignal, os.Interrupt, syscall.SIGTERM)
	<-cleanupSignal
	signal.Stop(cleanupSignal)

	fmt.Println("\n  Ending session. Closing tunnel and local server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Local server shutdown error: %v", err)
	}
	tunnel.Stop(tunnelCmd)
	fmt.Println("  Done.")
}

func buildShareLink(tunnelURL string, publicKey, token []byte) string {
	return fmt.Sprintf(
		"%s#u=%s&k=%s&t=%s",
		connectPageURL,
		base64.RawURLEncoding.EncodeToString([]byte(tunnelURL)),
		base64.RawURLEncoding.EncodeToString(publicKey),
		base64.RawURLEncoding.EncodeToString(token),
	)
}

func printShareLink(shareLink string) {
	fmt.Println()
	fmt.Println("  Skyping agent running")
	fmt.Println("  ========================================")
	fmt.Println("  Share this one-time encrypted link:")
	fmt.Println()
	fmt.Printf("  \033[1;32m%s\033[0m\n", shareLink)
	fmt.Println()
	fmt.Println("  Anyone without this exact link is rejected.")
	fmt.Println("  Press Ctrl+C to stop and remove access.")
	fmt.Println("  ========================================")
	fmt.Println()
}

func (s *session) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	if !atomic.CompareAndSwapInt32(&s.active, 0, 1) {
		_ = conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "Another session is already active"),
			time.Now().Add(writeTimeout),
		)
		return
	}
	defer atomic.StoreInt32(&s.active, 0)

	secret, err := s.performHandshake(conn)
	if err != nil {
		log.Printf("Viewer handshake rejected: %v", err)
		return
	}

	fmt.Println("  Viewer verified. End-to-end encryption active.")
	handleSessionEncrypted(conn, secret)
	fmt.Println("  Viewer disconnected.")
}

func (s *session) performHandshake(conn *websocket.Conn) ([]byte, error) {
	if err := conn.SetReadDeadline(time.Now().Add(handshakeTimeout)); err != nil {
		return nil, err
	}
	defer conn.SetReadDeadline(time.Time{})

	messageType, payload, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("read client hello: %w", err)
	}
	if messageType != websocket.TextMessage {
		return nil, fmt.Errorf("client hello must be text")
	}

	var hello clientHello
	if err := json.Unmarshal(payload, &hello); err != nil {
		return nil, fmt.Errorf("decode client hello: %w", err)
	}

	clientPublicKey, err := base64.RawURLEncoding.DecodeString(hello.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("decode viewer key: %w", err)
	}
	clientKey, err := ecdh.P256().NewPublicKey(clientPublicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid viewer key: %w", err)
	}

	proof, err := base64.RawURLEncoding.DecodeString(hello.Proof)
	if err != nil {
		return nil, fmt.Errorf("decode client proof: %w", err)
	}
	if !hmac.Equal(proof, handshakeProof(s.token, "client", clientPublicKey, s.publicKey)) {
		return nil, fmt.Errorf("invalid session secret")
	}

	secret, err := s.privateKey.ECDH(clientKey)
	if err != nil {
		return nil, fmt.Errorf("derive session key: %w", err)
	}

	serverProof := base64.RawURLEncoding.EncodeToString(handshakeProof(s.token, "server", clientPublicKey, s.publicKey))
	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return nil, err
	}
	defer conn.SetWriteDeadline(time.Time{})
	if err := conn.WriteJSON(serverHello{Proof: serverProof}); err != nil {
		return nil, fmt.Errorf("write server hello: %w", err)
	}

	return secret, nil
}

func handshakeProof(token []byte, role string, clientPublicKey, agentPublicKey []byte) []byte {
	mac := hmac.New(sha256.New, token)
	mac.Write([]byte("skyping/v1/"))
	mac.Write([]byte(role))
	mac.Write(clientPublicKey)
	mac.Write(agentPublicKey)
	return mac.Sum(nil)
}

func handleSessionEncrypted(conn *websocket.Conn, secret []byte) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	fmt.Printf("  Starting shell: %s\n", shell)
	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color", "COLUMNS=220", "LINES=50")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Printf("Failed to start shell: %v", err)
		return
	}
	defer ptmx.Close()
	defer stopProcessGroup(cmd)

	if err := pty.Setsize(ptmx, &pty.Winsize{Rows: 50, Cols: 220}); err != nil {
		log.Printf("Failed to set terminal size: %v", err)
	}
	fmt.Println("  Shell started. Bridging terminal...")

	done := make(chan struct{})
	var closeOnce sync.Once
	closeTransport := func() {
		closeOnce.Do(func() {
			_ = conn.Close()
			_ = ptmx.Close()
			close(done)
		})
	}

	go func() {
		defer closeTransport()
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				return
			}
			payload, err := encryptAESGCM(secret, buf[:n])
			if err != nil {
				log.Printf("Encrypt terminal output: %v", err)
				return
			}
			if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
				return
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, payload); err != nil {
				return
			}
		}
	}()

	go func() {
		defer closeTransport()
		for {
			messageType, payload, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if messageType != websocket.BinaryMessage {
				log.Printf("Ignoring unexpected WebSocket message type %d", messageType)
				continue
			}
			input, err := decryptAESGCM(secret, payload)
			if err != nil {
				log.Printf("Decrypt terminal input: %v", err)
				return
			}
			if _, err := ptmx.Write(input); err != nil {
				return
			}
		}
	}()

	<-done
	fmt.Println("  Shell exited.")
}

func stopProcessGroup(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM); err != nil {
		_ = cmd.Process.Kill()
	}
	_ = cmd.Wait()
}

func encryptAESGCM(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return aesgcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decryptAESGCM(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < aesgcm.NonceSize()+aesgcm.Overhead() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	return aesgcm.Open(nil, ciphertext[:aesgcm.NonceSize()], ciphertext[aesgcm.NonceSize():], nil)
}
