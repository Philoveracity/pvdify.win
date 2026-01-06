package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/philoveracity/pvdify/internal/client"
	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:     "apps",
	Aliases: []string{"app"},
	Short:   "List all apps",
	RunE:    runListApps,
}

var appsCreateCmd = &cobra.Command{
	Use:   "apps:create NAME",
	Short: "Create a new app",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreateApp,
}

var appsDeleteCmd = &cobra.Command{
	Use:     "apps:delete NAME",
	Aliases: []string{"apps:destroy"},
	Short:   "Delete an app",
	Args:    cobra.ExactArgs(1),
	RunE:    runDeleteApp,
}

var appsInfoCmd = &cobra.Command{
	Use:   "apps:info NAME",
	Short: "Show app details",
	Args:  cobra.ExactArgs(1),
	RunE:  runAppInfo,
}

var appEnv string

func init() {
	appsCreateCmd.Flags().StringVarP(&appEnv, "environment", "e", "production", "Environment (production, staging)")

	rootCmd.AddCommand(appsCreateCmd)
	rootCmd.AddCommand(appsDeleteCmd)
	rootCmd.AddCommand(appsInfoCmd)
}

func getClient() *client.Client {
	return client.New(apiURL, authToken)
}

func runListApps(cmd *cobra.Command, args []string) error {
	c := getClient()
	apps, err := c.ListApps()
	if err != nil {
		return err
	}

	if len(apps) == 0 {
		fmt.Println("No apps found. Create one with: pvdify apps:create NAME")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tIMAGE\tDOMAINS")
	for _, app := range apps {
		domains := "-"
		if len(app.Domains) > 0 {
			domains = app.Domains[0]
			if len(app.Domains) > 1 {
				domains += fmt.Sprintf(" (+%d)", len(app.Domains)-1)
			}
		}
		image := app.Image
		if image == "" {
			image = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", app.Name, app.Status, image, domains)
	}
	w.Flush()
	return nil
}

func runCreateApp(cmd *cobra.Command, args []string) error {
	name := args[0]
	c := getClient()

	app, err := c.CreateApp(name, appEnv)
	if err != nil {
		return err
	}

	fmt.Printf("Created app %s\n", app.Name)
	fmt.Printf("  Environment: %s\n", app.Environment)
	fmt.Printf("  Status: %s\n", app.Status)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  pvdify config:set %s KEY=VALUE  # Set environment variables\n", name)
	fmt.Printf("  pvdify deploy %s --image IMG    # Deploy a container image\n", name)
	return nil
}

func runDeleteApp(cmd *cobra.Command, args []string) error {
	name := args[0]
	c := getClient()

	if err := c.DeleteApp(name); err != nil {
		return err
	}

	fmt.Printf("Deleted app %s\n", name)
	return nil
}

func runAppInfo(cmd *cobra.Command, args []string) error {
	name := args[0]
	c := getClient()

	app, err := c.GetApp(name)
	if err != nil {
		return err
	}

	fmt.Printf("=== %s ===\n", app.Name)
	fmt.Printf("Environment: %s\n", app.Environment)
	fmt.Printf("Status: %s\n", app.Status)
	if app.Image != "" {
		fmt.Printf("Image: %s\n", app.Image)
	}
	if app.BindPort > 0 {
		fmt.Printf("Port: %d\n", app.BindPort)
	}
	if len(app.Domains) > 0 {
		fmt.Printf("Domains:\n")
		for _, d := range app.Domains {
			fmt.Printf("  - %s\n", d)
		}
	}
	fmt.Printf("Created: %s\n", app.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", app.UpdatedAt.Format("2006-01-02 15:04:05"))

	return nil
}
