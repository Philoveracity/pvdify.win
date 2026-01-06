package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var domainsCmd = &cobra.Command{
	Use:     "domains NAME",
	Aliases: []string{"domain"},
	Short:   "List domains for an app",
	Args:    cobra.ExactArgs(1),
	RunE:    runListDomains,
}

var domainsAddCmd = &cobra.Command{
	Use:   "domains:add NAME DOMAIN",
	Short: "Add a domain to an app",
	Args:  cobra.ExactArgs(2),
	RunE:  runAddDomain,
}

var domainsRemoveCmd = &cobra.Command{
	Use:     "domains:remove NAME DOMAIN",
	Aliases: []string{"domains:delete"},
	Short:   "Remove a domain from an app",
	Args:    cobra.ExactArgs(2),
	RunE:    runRemoveDomain,
}

func init() {
	rootCmd.AddCommand(domainsAddCmd)
	rootCmd.AddCommand(domainsRemoveCmd)
}

func runListDomains(cmd *cobra.Command, args []string) error {
	name := args[0]
	c := getClient()

	domains, err := c.ListDomains(name)
	if err != nil {
		return err
	}

	if len(domains) == 0 {
		fmt.Printf("No domains configured for %s\n", name)
		fmt.Printf("\nAdd one with: pvdify domains:add %s DOMAIN\n", name)
		return nil
	}

	fmt.Printf("=== %s Domains ===\n", name)
	for _, d := range domains {
		fmt.Printf("  %s\n", d)
	}
	return nil
}

func runAddDomain(cmd *cobra.Command, args []string) error {
	name := args[0]
	domain := args[1]
	c := getClient()

	if err := c.AddDomain(name, domain); err != nil {
		return err
	}

	fmt.Printf("Added %s to %s\n", domain, name)
	fmt.Printf("\nConfigure your DNS:\n")
	fmt.Printf("  %s CNAME -> %s.pvdify.win\n", domain, name)
	return nil
}

func runRemoveDomain(cmd *cobra.Command, args []string) error {
	name := args[0]
	domain := args[1]
	c := getClient()

	if err := c.RemoveDomain(name, domain); err != nil {
		return err
	}

	fmt.Printf("Removed %s from %s\n", domain, name)
	return nil
}
