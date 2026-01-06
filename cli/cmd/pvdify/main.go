package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version   = "0.1.0"
	apiURL    string
	authToken string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:     "pvdify",
	Short:   "Pvdify CLI - Heroku-style PaaS for AlmaLinux",
	Long:    `Pvdify provides Heroku-style deployment management for containerized applications.`,
	Version: version,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", getEnvOrDefault("PVDIFY_API_URL", "https://pvdify.win"), "Pvdify API URL")
	rootCmd.PersistentFlags().StringVar(&authToken, "token", os.Getenv("PVDIFY_TOKEN"), "API authentication token")

	// Add commands
	rootCmd.AddCommand(appsCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(releasesCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(domainsCmd)
	rootCmd.AddCommand(psCmd)
	rootCmd.AddCommand(logsCmd)
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
