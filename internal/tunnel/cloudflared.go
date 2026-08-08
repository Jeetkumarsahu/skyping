package tunnel

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"syscall"
	"time"
)

const startupTimeout = 30 * time.Second

var quickTunnelURL = regexp.MustCompile(`https://[a-z0-9-]+\.trycloudflare\.com`)

type startupResult struct {
	url string
	err error
}

func StartTunnel(port int) (string, *exec.Cmd, error) {
	binary := "cloudflared"
	if _, err := os.Stat("./cloudflared"); err == nil {
		binary = "./cloudflared"
	}

	cmd := exec.Command(binary, "tunnel", "--url", fmt.Sprintf("http://127.0.0.1:%d", port), "--no-autoupdate")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", nil, err
	}
	if err := cmd.Start(); err != nil {
		return "", nil, fmt.Errorf("failed to start cloudflared (%s): %w", binary, err)
	}

	result := make(chan startupResult, 1)
	go watchOutput(stderr, result)

	select {
	case startup := <-result:
		if startup.err != nil {
			Stop(cmd)
			return "", nil, startup.err
		}
		return startup.url, cmd, nil
	case <-time.After(startupTimeout):
		Stop(cmd)
		return "", nil, fmt.Errorf("cloudflared did not provide a quick-tunnel URL within %s", startupTimeout)
	}
}

func watchOutput(stderr io.Reader, result chan<- startupResult) {
	defer close(result)
	scanner := bufio.NewScanner(stderr)
	reported := false
	for scanner.Scan() {
		if !reported {
			if url := quickTunnelURL.FindString(scanner.Text()); url != "" {
				result <- startupResult{url: url}
				reported = true
				return
			}
		}
	}
	if !reported {
		if err := scanner.Err(); err != nil {
			result <- startupResult{err: fmt.Errorf("read cloudflared output: %w", err)}
			return
		}
		result <- startupResult{err: fmt.Errorf("cloudflared exited before creating a quick tunnel")}
	}
}

func Stop(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		<-done
	}
}
