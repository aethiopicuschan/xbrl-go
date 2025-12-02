package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aethiopicuschan/xbrl-go/pkg/xbrl"
)

var (
	conceptLocal    string
	conceptURI      string
	contextID       string
	unitID          string
	onlyNil         bool
	excludeNil      bool
	normalizeSpaces bool
)

var factsCmd = &cobra.Command{
	Use:   "facts <instance.xbrl>",
	Short: "List facts from an XBRL instance document",
	Long: `List facts from an XBRL instance document.

You can filter facts by concept, context ID, unit ID, and nil-state.

Examples:

  # List all facts
  xbrl-go facts sample.xbrl

  # List all facts with local name 'Revenue'
  xbrl-go facts --concept-local Revenue sample.xbrl

  # List facts in context C1
  xbrl-go facts --context C1 sample.xbrl

  # List non-nil Revenue facts in unit U1
  xbrl-go facts --concept-local Revenue --unit U1 --exclude-nil sample.xbrl
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if onlyNil && excludeNil {
			return fmt.Errorf("--only-nil and --exclude-nil cannot be used together")
		}

		path := args[0]

		doc, err := xbrl.ParseFile(path)
		if err != nil {
			return fmt.Errorf("parse instance: %w", err)
		}

		// Build filter
		filter := xbrl.NewFactFilter().
			ConceptLocal(conceptLocal).
			ConceptURI(conceptURI).
			ContextID(contextID).
			UnitID(unitID)

		if onlyNil {
			filter = filter.OnlyNil()
		} else if excludeNil {
			filter = filter.ExcludeNil()
		}

		facts := doc.Facts()
		if conceptLocal != "" || conceptURI != "" || contextID != "" || unitID != "" || onlyNil || excludeNil {
			facts = doc.FilterFacts(filter)
		}

		if len(facts) == 0 {
			fmt.Println("no facts matched the filter")
			return nil
		}

		fmt.Println("---- facts ----")
		for _, f := range facts {
			if f == nil {
				continue
			}

			name := f.Name().String()

			value := f.Value()
			if normalizeSpaces {
				// Use normalized value for human-readable output.
				value = f.NormalizedValue()
			}

			if f.IsNil() {
				value = "(nil)"
			}

			fmt.Printf(
				"%s\tctx=%s\tunit=%s\tdecimals=%s\tvalue=%s\n",
				name,
				f.ContextRef(),
				f.UnitRef(),
				f.Decimals(),
				value,
			)
		}

		return nil
	},
}

func init() {
	// Register subcommand on the root command.
	rootCmd.AddCommand(factsCmd)

	// Add flags to the facts command.
	factsCmd.Flags().StringVar(&conceptLocal, "concept-local", "", "filter facts by concept local name")
	factsCmd.Flags().StringVar(&conceptURI, "concept-uri", "", "filter facts by concept namespace URI")
	factsCmd.Flags().StringVar(&contextID, "context", "", "filter facts by context ID (contextRef)")
	factsCmd.Flags().StringVar(&unitID, "unit", "", "filter facts by unit ID (unitRef)")
	factsCmd.Flags().BoolVar(&onlyNil, "only-nil", false, "filter only nil facts (xsi:nil=\"true\")")
	factsCmd.Flags().BoolVar(&excludeNil, "exclude-nil", false, "filter only non-nil facts (xsi:nil!=\"true\")")
	factsCmd.Flags().BoolVar(&normalizeSpaces, "normalize-spaces", false, "normalize spaces in fact values for human-readable output")
}
