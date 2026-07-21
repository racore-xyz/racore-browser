package kubo

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Manager struct {
	apiURL     string
	gatewayURL string
	repoPath   string
	cmd        *exec.Cmd
	execPath   string
}

func New(apiURL, gatewayURL, dataDir string) *Manager {
	return &Manager{
		apiURL:     apiURL,
		gatewayURL: gatewayURL,
		repoPath:   filepath.Join(dataDir, "ipfs"),
	}
}

func (m *Manager) findExecutable() string {
	if v := os.Getenv("RACORE_KUBO_PATH"); v != "" {
		if _, err := os.Stat(v); err == nil {
			return v
		}
	}
	if p, err := exec.LookPath("ipfs"); err == nil {
		return p
	}
	return ""
}

func (m *Manager) isOnline(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "POST", m.apiURL+"/api/v0/id", nil)
	if err != nil {
		return false
	}
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

func (m *Manager) Start(ctx context.Context) (map[string]any, error) {
	if m.isOnline(ctx) {
		return map[string]any{"managed": false, "online": true, "reason": "external-node"}, nil
	}

	m.execPath = m.findExecutable()
	if m.execPath == "" {
		return map[string]any{"managed": false, "online": false, "reason": "kubo-not-installed"}, nil
	}

	os.MkdirAll(m.repoPath, 0700)
	env := append(os.Environ(), "IPFS_PATH="+m.repoPath)

	if _, err := os.Stat(filepath.Join(m.repoPath, "config")); err != nil {
		initCmd := exec.CommandContext(ctx, m.execPath, "init", "--profile=lowpower")
		initCmd.Env = env
		if out, err := initCmd.CombinedOutput(); err != nil {
			return map[string]any{"managed": false, "online": false, "reason": fmt.Sprintf("kubo-init-failed: %s", string(out))}, nil
		}
	}

	gatewayPort := "8180"
	if n, err := fmt.Sscanf(m.gatewayURL, "%*s%d", &gatewayPort); err != nil && n != 1 {
		gatewayPort = "8180"
	}

	configCmd := exec.CommandContext(ctx, m.execPath, "config", "Addresses.Gateway", fmt.Sprintf("/ip4/127.0.0.1/tcp/%s", gatewayPort))
	configCmd.Env = env
	configCmd.Run()

	m.cmd = exec.CommandContext(ctx, m.execPath, "daemon", "--enable-gc")
	m.cmd.Env = env
	m.cmd.Stdout = nil
	m.cmd.Stderr = nil

	if err := m.cmd.Start(); err != nil {
		return map[string]any{"managed": true, "online": false, "reason": fmt.Sprintf("kubo-daemon-start: %v", err)}, nil
	}

	for range 40 {
		if m.isOnline(ctx) {
			return map[string]any{"managed": true, "online": true, "pid": m.cmd.Process.Pid}, nil
		}
		if m.cmd.ProcessState != nil && m.cmd.ProcessState.Exited() {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	m.Stop()
	return map[string]any{"managed": true, "online": false, "reason": "kubo-start-timeout"}, nil
}

func (m *Manager) Stop() {
	if m.cmd != nil && m.cmd.Process != nil {
		m.cmd.Process.Signal(os.Interrupt)
		done := make(chan any, 1)
		go func() {
			m.cmd.Wait()
			done <- nil
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			m.cmd.Process.Kill()
		}
	}
}
