// Package hostctl implementa el cliente que habla con el helper hostctl (el daemon que
// corre directo en el host via systemd, ver /hostctl en la raiz del repo y CLAUDE.md).
// Hasta que ese daemon se instale (Fase 6), el socket simplemente no existe todavia y
// todas las llamadas fallan con un error claro -- eso se maneja arriba, en
// internal/metrics/throttle.go, degradando con gracia en vez de romper el dashboard.
package hostctl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type Client struct {
	socketPath string
	token      string
	timeout    time.Duration
}

func NewClient(socketPath, tokenFile string) *Client {
	token := ""
	if tokenFile != "" {
		if data, err := os.ReadFile(tokenFile); err == nil {
			token = strings.TrimSpace(string(data))
		}
	}
	return &Client{socketPath: socketPath, token: token, timeout: 2 * time.Second}
}

type request struct {
	Cmd   string `json:"cmd"`
	Token string `json:"token,omitempty"`
}

type response struct {
	OK    bool            `json:"ok"`
	Error string          `json:"error,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

func (c *Client) call(cmd string) (json.RawMessage, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("hostctl no disponible: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(c.timeout))

	if err := json.NewEncoder(conn).Encode(request{Cmd: cmd, Token: c.token}); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("hostctl: respuesta vacia")
	}

	var resp response
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("hostctl: %s", resp.Error)
	}
	return resp.Data, nil
}

type ThrottledStatus struct {
	Raw string `json:"raw"`
}

func (c *Client) GetThrottled() (ThrottledStatus, error) {
	data, err := c.call("get_throttled")
	if err != nil {
		return ThrottledStatus{}, err
	}
	var out ThrottledStatus
	if err := json.Unmarshal(data, &out); err != nil {
		return ThrottledStatus{}, err
	}
	return out, nil
}

func (c *Client) Reboot() error {
	_, err := c.call("reboot")
	return err
}

func (c *Client) Poweroff() error {
	_, err := c.call("poweroff")
	return err
}

func (c *Client) Ping() error {
	_, err := c.call("ping")
	return err
}
