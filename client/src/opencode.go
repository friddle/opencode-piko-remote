package src

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"time"
)

type OpencodeProcess struct {
	config  *Config
	binPath string
	port    int
	cmd     *exec.Cmd
}

func NewOpencodeProcess(config *Config, binPath string, port int) *OpencodeProcess {
	return &OpencodeProcess{
		config:  config,
		binPath: binPath,
		port:    port,
	}
}

func (o *OpencodeProcess) Start(ctx context.Context) error {
	args := []string{
		"web",
		"--port", fmt.Sprintf("%d", o.port),
		"--hostname", "127.0.0.1",
		"--cors", o.config.GetCorsOrigin(),
	}

	if o.config.Project != "" {
		args = append(args, o.config.Project)
	}

	o.cmd = exec.CommandContext(ctx, o.binPath, args...)

	env := o.cmd.Environ()
	if o.config.Pass != "" {
		env = append(env, fmt.Sprintf("OPENCODE_SERVER_PASSWORD=%s", o.config.Pass))
	}
	if o.config.User != "" {
		env = append(env, fmt.Sprintf("OPENCODE_SERVER_USERNAME=%s", o.config.User))
	}
	env = append(env, fmt.Sprintf("OPENCODE_WEB_BASE=/%s", o.config.Name))
	o.cmd.Env = env

	if err := o.cmd.Start(); err != nil {
		return fmt.Errorf("start opencode: %w", err)
	}

	if err := o.waitReady(ctx); err != nil {
		return fmt.Errorf("opencode not ready: %w", err)
	}

	return nil
}

func (o *OpencodeProcess) Wait() error {
	if o.cmd != nil {
		return o.cmd.Wait()
	}
	return nil
}

func (o *OpencodeProcess) Stop() {
	if o.cmd != nil && o.cmd.Process != nil {
		o.cmd.Process.Kill()
	}
}

func (o *OpencodeProcess) Port() int {
	return o.port
}

func (o *OpencodeProcess) waitReady(ctx context.Context) error {
	url := fmt.Sprintf("http://127.0.0.1:%d/", o.port)
	client := &http.Client{Timeout: 2 * time.Second}

	for i := 0; i < 30; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 || resp.StatusCode == 401 {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for opencode on port %d", o.port)
}
