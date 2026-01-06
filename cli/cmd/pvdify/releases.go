package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var deployImage string

var deployCmd = &cobra.Command{
	Use:   "deploy NAME",
	Short: "Deploy a container image to an app",
	Args:  cobra.ExactArgs(1),
	RunE:  runDeploy,
}

var releasesCmd = &cobra.Command{
	Use:     "releases NAME",
	Aliases: []string{"release"},
	Short:   "List releases for an app",
	Args:    cobra.ExactArgs(1),
	RunE:    runListReleases,
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback NAME",
	Short: "Rollback to the previous release",
	Args:  cobra.ExactArgs(1),
	RunE:  runRollback,
}

func init() {
	deployCmd.Flags().StringVarP(&deployImage, "image", "i", "", "Container image to deploy (required)")
	deployCmd.MarkFlagRequired("image")
}

func runDeploy(cmd *cobra.Command, args []string) error {
	name := args[0]
	c := getClient()

	fmt.Printf("Deploying %s to %s...\n", deployImage, name)

	release, err := c.CreateRelease(name, deployImage)
	if err != nil {
		return err
	}

	fmt.Printf("Released v%d\n", release.Version)
	fmt.Printf("  Image: %s\n", release.Image)
	fmt.Printf("  Status: %s\n", release.Status)
	return nil
}

func runListReleases(cmd *cobra.Command, args []string) error {
	name := args[0]
	c := getClient()

	releases, err := c.ListReleases(name)
	if err != nil {
		return err
	}

	if len(releases) == 0 {
		fmt.Printf("No releases found for %s\n", name)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "VERSION\tIMAGE\tSTATUS\tCREATED")
	for _, r := range releases {
		fmt.Fprintf(w, "v%d\t%s\t%s\t%s\n",
			r.Version,
			truncate(r.Image, 50),
			r.Status,
			r.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}
	w.Flush()
	return nil
}

func runRollback(cmd *cobra.Command, args []string) error {
	name := args[0]
	c := getClient()

	fmt.Printf("Rolling back %s...\n", name)

	release, err := c.Rollback(name)
	if err != nil {
		return err
	}

	fmt.Printf("Rolled back to v%d\n", release.Version)
	fmt.Printf("  Image: %s\n", release.Image)
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
