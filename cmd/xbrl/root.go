package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/aethiopicuschan/xbrl-go/pkg/xbrl"
)

var rootCmd = &cobra.Command{
	Use:   "xbrl <instance.xbrl>",
	Short: "xbrl is a CLI for working with XBRL instance documents",
	Long: `xbrl is a CLI tool built on top of the xbrl-go library.

By default it prints a summary of the instance document:
  - number of schemaRefs
  - number of contexts
  - number of units
  - number of facts

Use the 'facts' subcommand to inspect individual facts with filters.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		doc, err := xbrl.ParseFile(path)
		if err != nil {
			return fmt.Errorf("parse instance: %w", err)
		}

		fmt.Printf("schemaRefs: %d\n", len(doc.SchemaRefs()))
		fmt.Printf("contexts  : %d\n", len(doc.Contexts()))
		fmt.Printf("units     : %d\n", len(doc.Units()))
		fmt.Printf("facts     : %d\n", len(doc.Facts()))

		return nil
	},
}

func init() {
	bi, ok := debug.ReadBuildInfo()
	if ok {
		rootCmd.Version = bi.Main.Version
	}
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
