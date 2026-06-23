package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

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

func daemonize(cmd *cobra.Command, args []string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find executable: %w", err)
	}

	if pid, err := readPID(); err == nil && isRunning(pid) {
		return fmt.Errorf("daemon already running (PID %d), use 'opencode-piko stop' first", pid)
	}

	logFile, err := os.OpenFile(logFilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	daemonArgs := []string{"--foreground"}
	daemonArgs = append(daemonArgs, args...)
	daemonArgs = append(daemonArgs, fmt.Sprintf("--name=%s", nameFlag))
	daemonArgs = append(daemonArgs, fmt.Sprintf("--remote=%s", remoteFlag))
	daemonArgs = append(daemonArgs, fmt.Sprintf("--pass=%s", passFlag))
	daemonArgs = append(daemonArgs, fmt.Sprintf("--user=%s", userFlag))
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
	fmt.Printf("  Log: %s\n", logFilePath())
	fmt.Printf("  Stop: opencode-piko stop\n")
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
