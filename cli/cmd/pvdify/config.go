package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config NAME",
	Short: "Show config vars for an app",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetConfig,
}

var configSetCmd = &cobra.Command{
	Use:   "config:set NAME KEY=VALUE [KEY=VALUE...]",
	Short: "Set config vars",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runSetConfig,
}

var configUnsetCmd = &cobra.Command{
	Use:   "config:unset NAME KEY [KEY...]",
	Short: "Unset config vars",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runUnsetConfig,
}

func init() {
	rootCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configUnsetCmd)
}

func runGetConfig(cmd *cobra.Command, args []string) error {
	name := args[0]
	c := getClient()

	config, err := c.GetConfig(name)
	if err != nil {
		return err
	}

	if len(config) == 0 {
		fmt.Printf("No config vars set for %s\n", name)
		return nil
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Printf("=== %s Config Vars ===\n", name)
	for _, k := range keys {
		fmt.Printf("%s: %s\n", k, config[k])
	}
	return nil
}

func runSetConfig(cmd *cobra.Command, args []string) error {
	name := args[0]
	c := getClient()

	config := make(map[string]string)
	for _, kv := range args[1:] {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid config format: %s (expected KEY=VALUE)", kv)
		}
		config[parts[0]] = parts[1]
	}

	if err := c.SetConfig(name, config); err != nil {
		return err
	}

	fmt.Printf("Setting config vars and restarting %s...\n", name)
	for k, v := range config {
		fmt.Printf("  %s: %s\n", k, v)
	}
	fmt.Println("done")
	return nil
}

func runUnsetConfig(cmd *cobra.Command, args []string) error {
	name := args[0]
	c := getClient()

	for _, key := range args[1:] {
		if err := c.UnsetConfig(name, key); err != nil {
			return fmt.Errorf("failed to unset %s: %w", key, err)
		}
		fmt.Printf("Unsetting %s...\n", key)
	}
	fmt.Printf("Restarting %s... done\n", name)
	return nil
}
