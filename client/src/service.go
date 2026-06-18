package src

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/andydunstall/piko/agent/config"
	"github.com/andydunstall/piko/agent/reverseproxy"
	"github.com/andydunstall/piko/client"
	"github.com/andydunstall/piko/pkg/log"
	"github.com/oklog/run"
)

type ServiceManager struct {
	config *Config
	ctx    context.Context
	cancel context.CancelFunc
}

func NewServiceManager(config *Config) *ServiceManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ServiceManager{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (sm *ServiceManager) Start() error {
	binPath, err := EnsureOpencode()
	if err != nil {
		return fmt.Errorf("ensure opencode: %w", err)
	}

	ocPort := FindAvailablePort()
	fmt.Printf("Starting opencode-piko\n")
	fmt.Printf("  Name:    %s\n", sm.config.Name)
	fmt.Printf("  Remote:  %s\n", sm.config.Remote)
	fmt.Printf("  Port:    %d\n", ocPort)
	if sm.config.Project != "" {
		fmt.Printf("  Project: %s\n", sm.config.Project)
	}

	oc := NewOpencodeProcess(sm.config, binPath, ocPort)
	proxyPort := FindAvailablePort()

	rp, err := NewRewriteProxy(
		fmt.Sprintf("http://127.0.0.1:%d", ocPort),
		sm.config.Name,
	)
	if err != nil {
		return fmt.Errorf("create rewrite proxy: %w", err)
	}

	var g run.Group

	// opencode web process
	g.Add(func() error {
		if err := oc.Start(sm.ctx); err != nil {
			return err
		}
		fmt.Printf("opencode web started on port %d\n", ocPort)
		return oc.Wait()
	}, func(error) {
		oc.Stop()
	})

	// rewrite proxy
	proxySrv := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", proxyPort),
		Handler: rp.Handler(),
	}
	g.Add(func() error {
		fmt.Printf("rewrite proxy on port %d -> %d\n", proxyPort, ocPort)
		go func() {
			<-sm.ctx.Done()
			proxySrv.Close()
		}()
		return proxySrv.ListenAndServe()
	}, func(error) {
		proxySrv.Close()
	})

	// piko agent (connects to proxy, not directly to opencode)
	g.Add(func() error {
		return sm.startPiko(proxyPort)
	}, func(error) {
		sm.cancel()
	})

	// signal handler
	g.Add(func() error {
		c := make(chan os.Signal, 1)
		if runtime.GOOS == "windows" {
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		} else {
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		}
		select {
		case sig := <-c:
			fmt.Printf("\nReceived signal %v, shutting down...\n", sig)
			sm.cancel()
			return nil
		case <-sm.ctx.Done():
			return sm.ctx.Err()
		}
	}, func(error) {
		sm.cancel()
	})

	// auto exit timer
	if sm.config.AutoExit {
		g.Add(func() error {
			timer := time.NewTimer(24 * time.Hour)
			defer timer.Stop()
			select {
			case <-timer.C:
				fmt.Printf("\n24h auto-exit triggered\n")
				sm.cancel()
				return nil
			case <-sm.ctx.Done():
				return sm.ctx.Err()
			}
		}, func(error) {
			sm.cancel()
		})
	}

	if sm.config.Pass != "" {
		fmt.Printf("  Auth:    %s / %s\n", sm.config.User, sm.config.Pass)
	} else {
		fmt.Printf("  Auth:    disabled\n")
	}
	fmt.Printf("Press Ctrl+C to stop\n")

	return g.Run()
}

func (sm *ServiceManager) startPiko(localPort int) error {
	remote := sm.config.Remote
	if !strings.HasPrefix(remote, "http") {
		remote = fmt.Sprintf("http://%s", remote)
	}

	conf := &config.Config{
		Connect: config.ConnectConfig{
			URL:     remote,
			Timeout: 30 * time.Second,
		},
		Listeners: []config.ListenerConfig{
			{
				EndpointID: sm.config.Name,
				Protocol:   config.ListenerProtocolHTTP,
				Addr:       fmt.Sprintf("127.0.0.1:%d", localPort),
				AccessLog:  false,
				Timeout:    30 * time.Second,
			},
		},
		Log: log.Config{
			Level:      "info",
			Subsystems: []string{},
		},
		GracePeriod: 30 * time.Second,
	}

	logger, err := log.NewLogger("info", []string{})
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}

	if err := conf.Validate(); err != nil {
		return fmt.Errorf("piko config invalid: %w", err)
	}

	connectURL, err := url.Parse(conf.Connect.URL)
	if err != nil {
		return fmt.Errorf("parse connect URL: %w", err)
	}

	upstream := &client.Upstream{
		URL:    connectURL,
		Logger: logger.WithSubsystem("client"),
	}

	for _, listenerConfig := range conf.Listeners {
		ln, err := upstream.Listen(sm.ctx, listenerConfig.EndpointID)
		if err != nil {
			return fmt.Errorf("listen endpoint %s: %w", listenerConfig.EndpointID, err)
		}
		fmt.Printf("Connected to piko endpoint: %s\n", listenerConfig.EndpointID)

		metrics := reverseproxy.NewMetrics("proxy")
		server := reverseproxy.NewServer(listenerConfig, metrics, logger)
		if server == nil {
			return fmt.Errorf("create proxy server failed")
		}

		go func() {
			if err := server.Serve(ln); err != nil && err != context.Canceled {
				fmt.Printf("proxy error: %v\n", err)
			}
		}()
	}

	<-sm.ctx.Done()
	return sm.ctx.Err()
}
