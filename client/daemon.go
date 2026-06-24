package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"opencode-piko-remote/src"

	"github.com/spf13/cobra"
)

func daemonDir() string {
	dir := filepath.Join(os.Getenv("HOME"), ".opencode-piko")
	os.MkdirAll(dir, 0755)
	return dir
}

func pidFilePath() string {
	return filepath.Join(daemonDir(), "daemon.pid")
}

func logFilePath() string {
	return filepath.Join(daemonDir(), "daemon.log")
}

func writePID(pid int) error {
	return os.WriteFile(pidFilePath(), []byte(fmt.Sprintf("%d\n", pid)), 0644)
}

func readPID() (int, error) {
	data, err := os.ReadFile(pidFilePath())
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(string(data[:len(data)-1]))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func isRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}

func printAccessInfo(config *src.Config) {
	remote := config.Remote
	if len(remote) > 0 && remote[0] != ':' {
		parts := splitHostPort(remote)
		scheme := "https"
		fmt.Printf("  Remote:     %s://%s\n", scheme, remote)
		fmt.Printf("  Access URL: %s://%s/%s/\n", scheme, parts, config.Name)
	}
	if config.Pass != "" {
		fmt.Printf("  Auth:       %s / %s\n", config.User, config.Pass)
	} else {
		fmt.Printf("  Auth:       disabled\n")
	}
}

func splitHostPort(addr string) string {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
		if addr[i] < '0' || addr[i] > '9' {
			break
		}
	}
	return addr
}

func daemonize(cmd *cobra.Command, args []string) error {
	project := ""
	if len(args) > 0 {
		project = args[0]
	}

	config := &src.Config{
		Name:       nameFlag,
		Remote:     remoteFlag,
		ServerPort: serverPortFlag,
		User:       userFlag,
		Pass:       passFlag,
		Project:    project,
		AutoExit:   autoExitFlag,
	}
	config.Validate()

	if pid, err := readPID(); err == nil && isRunning(pid) {
		return fmt.Errorf("daemon already running (PID %d), use 'opencode-piko stop' first", pid)
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find executable: %w", err)
	}

	logFile, err := os.OpenFile(logFilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	daemonArgs := []string{}
	if project != "" {
		daemonArgs = append(daemonArgs, project)
	}
	daemonArgs = append(daemonArgs, fmt.Sprintf("--name=%s", config.Name))
	daemonArgs = append(daemonArgs, fmt.Sprintf("--remote=%s", config.Remote))
	daemonArgs = append(daemonArgs, fmt.Sprintf("--pass=%s", config.Pass))
	daemonArgs = append(daemonArgs, fmt.Sprintf("--user=%s", config.User))
	daemonArgs = append(daemonArgs, fmt.Sprintf("--server-port=%d", serverPortFlag))
	if autoExitFlag {
		daemonArgs = append(daemonArgs, "--auto-exit=true")
	} else {
		daemonArgs = append(daemonArgs, "--auto-exit=false")
	}

	c := exec.Command(execPath, daemonArgs...)
	c.Stdin = nil
	c.Stdout = logFile
	c.Stderr = logFile
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := c.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}

	pid := c.Process.Pid
	writePID(pid)

	fmt.Printf("opencode-piko daemon started (PID %d)\n", pid)
	fmt.Printf("  Name:       %s\n", config.Name)
	printAccessInfo(config)
	fmt.Printf("  Log:        %s\n", logFilePath())
	fmt.Printf("  Stop:       opencode-piko stop\n")
	return nil
}

func makeStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the running daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			pid, err := readPID()
			if err != nil {
				return fmt.Errorf("no daemon running")
			}
			if !isRunning(pid) {
				os.Remove(pidFilePath())
				return fmt.Errorf("daemon not running (stale PID)")
			}
			if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
				return fmt.Errorf("stop daemon: %w", err)
			}
			os.Remove(pidFilePath())
			fmt.Printf("opencode-piko daemon stopped (PID %d)\n", pid)
			return nil
		},
	}
}

func makeStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show daemon status",
		RunE: func(cmd *cobra.Command, args []string) error {
			pid, err := readPID()
			if err != nil || !isRunning(pid) {
				fmt.Println("opencode-piko daemon: not running")
				return nil
			}
			fmt.Printf("opencode-piko daemon: running (PID %d)\n", pid)
			fmt.Printf("  Log: %s\n", logFilePath())
			return nil
		},
	}
}
