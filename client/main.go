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

func main() {
	rootCmd := makeRootCmd()
	rootCmd.AddCommand(makeUpgradeCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func makeRootCmd() *cobra.Command {
	var (
		name       string
		remote     string
		serverPort int
		user       string
		pass       string
		autoExit   bool
	)

	cmd := &cobra.Command{
		Use:   "opencode-piko [project]",
		Short: "Expose opencode web through piko tunnel",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project := ""
			if len(args) > 0 {
				project = args[0]
			}

			config := &src.Config{
				Name:       name,
				Remote:     remote,
				ServerPort: serverPort,
				User:       user,
				Pass:       pass,
				Project:    project,
				AutoExit:   autoExit,
			}

			if err := config.Validate(); err != nil {
				return err
			}

			manager := src.NewServiceManager(config)
			return manager.Start()
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Piko endpoint name (required)")
	cmd.Flags().StringVar(&remote, "remote", "", "Piko server address host:port (required)")
	cmd.Flags().IntVar(&serverPort, "server-port", 8022, "Piko upstream port")
	cmd.Flags().StringVar(&user, "user", "opencode", "Auth username")
	cmd.Flags().StringVar(&pass, "pass", "", "Auth password")
	cmd.Flags().BoolVar(&autoExit, "auto-exit", true, "Auto exit after 24 hours")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("remote")

	cmd.Version = Version

	return cmd
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
