package main

import (
	"bufio"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var (
	logLines  int
	logFollow bool
)

var logsCmd = &cobra.Command{
	Use:     "logs NAME",
	Aliases: []string{"log"},
	Short:   "View logs for an app",
	Args:    cobra.ExactArgs(1),
	RunE:    runLogs,
}

func init() {
	logsCmd.Flags().IntVarP(&logLines, "lines", "n", 100, "Number of lines to show")
	logsCmd.Flags().BoolVarP(&logFollow, "follow", "f", false, "Follow log output")
}

func runLogs(cmd *cobra.Command, args []string) error {
	name := args[0]
	c := getClient()

	body, err := c.GetLogs(name, logLines, logFollow)
	if err != nil {
		return err
	}
	defer body.Close()

	reader := bufio.NewReader(body)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			fmt.Print(line)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	return nil
}
