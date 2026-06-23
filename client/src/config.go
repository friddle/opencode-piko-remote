package src

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const DefaultRemote = "clauded.friddle.me"

type Config struct {
	Name       string
	Remote     string
	ServerPort int
	User       string
	Pass       string
	Project    string
	AutoExit   bool
}

func (c *Config) ApplyDefaults() {
	if c.Remote == "" {
		c.Remote = DefaultRemote
	}
	if c.Name == "" {
		dir := c.Project
		if dir == "" {
			dir, _ = os.Getwd()
		}
		base := filepath.Base(dir)
		suffix := fmt.Sprintf("%04d", rand.Intn(10000))
		c.Name = base + "-" + suffix
	}
}

func (c *Config) Validate() error {
	c.ApplyDefaults()
	return nil
}

func (c *Config) GetRemoteHost() string {
	parts := strings.Split(c.Remote, ":")
	if len(parts) >= 1 {
		return parts[0]
	}
	return "localhost"
}

func (c *Config) GetRemotePort() int {
	parts := strings.Split(c.Remote, ":")
	if len(parts) >= 2 {
		if port, err := strconv.Atoi(parts[1]); err == nil {
			return port
		}
	}
	return 8088
}

func (c *Config) GetCorsOrigin() string {
	remote := c.Remote
	if strings.HasPrefix(remote, "http") {
		return remote
	}
	return fmt.Sprintf("http://%s", remote)
}

func FindAvailablePort() int {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 19876
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}
