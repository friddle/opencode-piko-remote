package main

import (
	"fmt"
	"os"

	"opencode-piko-remote/src"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	BuildTime = ""
	GitCommit = ""
)

var (
	nameFlag       string
	remoteFlag     string
	serverPortFlag int
	userFlag       string
	passFlag       string
	autoExitFlag   bool
	daemonFlag     bool
)

func main() {
	rootCmd := makeRootCmd()
	rootCmd.AddCommand(makeUpgradeCmd())
	rootCmd.AddCommand(makeStopCmd())
	rootCmd.AddCommand(makeStatusCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func makeRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "opencode-piko [project]",
		Short: "Expose opencode web through piko tunnel",
		Args:  cobra.MaximumNArgs(1),
		RunE:  run,
	}

	cmd.Flags().StringVar(&nameFlag, "name", "", "Piko endpoint name (default: <dir>-<random4>)")
	cmd.Flags().StringVar(&remoteFlag, "remote", "", "Piko server address host:port (default: clauded.friddle.me)")
	cmd.Flags().IntVar(&serverPortFlag, "server-port", 8022, "Piko upstream port")
	cmd.Flags().StringVar(&userFlag, "user", "opencode", "Auth username")
	cmd.Flags().StringVar(&passFlag, "pass", "", "Auth password")
	cmd.Flags().BoolVar(&autoExitFlag, "auto-exit", true, "Auto exit after 24 hours")
	cmd.Flags().BoolVarP(&daemonFlag, "daemon", "d", false, "Run as daemon (background)")

	cmd.Version = Version

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	project := ""
	if len(args) > 0 {
		project = args[0]
	}

	if daemonFlag {
		return daemonize(cmd, args)
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

	if err := config.Validate(); err != nil {
		return err
	}

	fmt.Printf("Starting opencode-piko\n")
	fmt.Printf("  Name:       %s\n", config.Name)
	printAccessInfo(config)
	fmt.Printf("  (Tip: use --daemon or -d to run in background)\n")

	manager := src.NewServiceManager(config)
	return manager.Start()
}

func makeUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade opencode to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return src.UpgradeOpencode()
		},
	}
}
