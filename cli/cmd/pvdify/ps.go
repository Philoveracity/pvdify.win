package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:     "ps NAME",
	Aliases: []string{"processes"},
	Short:   "List processes for an app",
	Args:    cobra.ExactArgs(1),
	RunE:    runListProcesses,
}

var psScaleCmd = &cobra.Command{
	Use:   "ps:scale NAME TYPE=COUNT [TYPE=COUNT...]",
	Short: "Scale process types",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runScale,
}

var psRestartCmd = &cobra.Command{
	Use:   "ps:restart NAME",
	Short: "Restart all processes",
	Args:  cobra.ExactArgs(1),
	RunE:  runRestart,
}

func init() {
	rootCmd.AddCommand(psScaleCmd)
	rootCmd.AddCommand(psRestartCmd)
}

func runListProcesses(cmd *cobra.Command, args []string) error {
	name := args[0]
	c := getClient()

	processes, err := c.ListProcesses(name)
	if err != nil {
		return err
	}

	if len(processes) == 0 {
		fmt.Printf("No processes running for %s\n", name)
		fmt.Printf("\nDeploy an image first: pvdify deploy %s --image IMAGE\n", name)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TYPE\tCOUNT\tCOMMAND\tSTATUS")
	for _, p := range processes {
		cmd := p.Command
		if cmd == "" {
			cmd = "-"
		}
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", p.Type, p.Count, cmd, p.Status)
	}
	w.Flush()
	return nil
}

func runScale(cmd *cobra.Command, args []string) error {
	name := args[0]
	c := getClient()

	processes := make(map[string]int)
	for _, spec := range args[1:] {
		parts := strings.SplitN(spec, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid scale format: %s (expected TYPE=COUNT)", spec)
		}
		count, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("invalid count for %s: %s", parts[0], parts[1])
		}
		if count < 0 {
			return fmt.Errorf("count cannot be negative: %s=%d", parts[0], count)
		}
		processes[parts[0]] = count
	}

	fmt.Printf("Scaling %s processes...\n", name)
	for t, c := range processes {
		fmt.Printf("  %s: %d\n", t, c)
	}

	if err := c.Scale(name, processes); err != nil {
		return err
	}

	fmt.Println("done")
	return nil
}

func runRestart(cmd *cobra.Command, args []string) error {
	name := args[0]
	c := getClient()

	fmt.Printf("Restarting %s...\n", name)

	if err := c.Restart(name); err != nil {
		return err
	}

	fmt.Println("done")
	return nil
}
