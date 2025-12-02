package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aethiopicuschan/xbrl-go/pkg/xbrl"
)

// FactDTO is an example DTO for exporting facts as JSON.
type FactDTO struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	ContextRef string `json:"context"`
	UnitRef    string `json:"unit"`
	Nil        bool   `json:"nil"`
}

// ExportFacts converts all facts in a Document into a slice of DTOs.
func ExportFacts(doc *xbrl.Document) []FactDTO {
	out := make([]FactDTO, 0, len(doc.Facts()))
	for _, f := range doc.Facts() {
		if f == nil {
			continue
		}

		value := f.NormalizedValue()
		if f.IsNil() {
			value = ""
		}

		out = append(out, FactDTO{
			Name:       f.Name().String(),
			Value:      value,
			ContextRef: f.ContextRef(),
			UnitRef:    f.UnitRef(),
			Nil:        f.IsNil(),
		})
	}
	return out
}

func main() {
	// Parse XBRL instance file
	doc, err := xbrl.ParseFile("sample.xbrl")
	if err != nil {
		log.Fatalf("failed to parse XBRL: %v", err)
	}

	// --- Summary ---
	fmt.Println("== Summary ==")
	fmt.Printf("schemaRefs: %d\n", len(doc.SchemaRefs()))
	fmt.Printf("contexts  : %d\n", len(doc.Contexts()))
	fmt.Printf("units     : %d\n", len(doc.Units()))
	fmt.Printf("facts     : %d\n", len(doc.Facts()))
	fmt.Println()

	// --- List all facts ---
	fmt.Println("== All facts ==")
	for _, f := range doc.Facts() {
		if f == nil {
			continue
		}
		name := f.Name().String()
		value := f.Value()
		if f.IsNil() {
			value = "(nil)"
		}
		fmt.Printf("%s  ctx=%s  unit=%s  decimals=%s  value=%s\n",
			name,
			f.ContextRef(),
			f.UnitRef(),
			f.Decimals(),
			value,
		)
	}
	fmt.Println()

	// --- Filter facts: example (concept local name = "Revenue", non-nil only) ---
	fmt.Println("== Filtered facts: conceptLocal=Revenue, non-nil ==")
	filter := xbrl.NewFactFilter().
		ConceptLocal("Revenue").
		ExcludeNil()

	filtered := doc.FilterFacts(filter)
	if len(filtered) == 0 {
		fmt.Println("no facts matched the filter")
	} else {
		for _, f := range filtered {
			if f == nil {
				continue
			}
			fmt.Printf("%s  ctx=%s  unit=%s  value=%s\n",
				f.Name().String(),
				f.ContextRef(),
				f.UnitRef(),
				f.Value(),
			)
		}
	}
	fmt.Println()

	// --- Inspect contexts ---
	fmt.Println("== Contexts ==")
	for id, ctx := range doc.Contexts() {
		fmt.Println("Context ID:", id)

		ent := ctx.Entity().Identifier()
		fmt.Printf("  Entity: %s (scheme=%s)\n", ent.Value(), ent.Scheme())

		p := ctx.Period()
		switch {
		case p.IsInstant():
			inst, _ := p.Instant()
			fmt.Printf("  Period: instant=%s\n", inst)
		case p.IsForever():
			fmt.Println("  Period: forever")
		default:
			start, _ := p.StartDate()
			end, _ := p.EndDate()
			fmt.Printf("  Period: %s to %s\n", start, end)
		}
	}
	fmt.Println()

	// --- Inspect units ---
	fmt.Println("== Units ==")
	for id, unit := range doc.Units() {
		fmt.Println("Unit ID:", id)
		if unit.IsDivide() {
			fmt.Println("  (divide unit)")
			for _, m := range unit.NumeratorMeasures() {
				fmt.Printf("  numerator: %s (prefix=%s, uri=%s)\n",
					m.Local(), m.Prefix(), m.URI())
			}
			for _, m := range unit.DenominatorMeasures() {
				fmt.Printf("  denominator: %s (prefix=%s, uri=%s)\n",
					m.Local(), m.Prefix(), m.URI())
			}
		} else {
			for _, m := range unit.Measures() {
				fmt.Printf("  measure: %s (prefix=%s, uri=%s)\n",
					m.Local(), m.Prefix(), m.URI())
			}
		}
	}
	fmt.Println()

	// --- Export to JSON (example) ---
	fmt.Println("== Facts as JSON ==")
	if err := doc.EncodeFactsJSON(os.Stdout, true); err != nil {
		log.Fatalf("failed to encode JSON: %v", err)
	}
	fmt.Println()

	// --- Working with Taxonomy and Concepts ---
	fmt.Println("== Taxonomy / Concepts ==")

	// Load taxonomy from XSD file
	tax, err := xbrl.ParseTaxonomyFile("sample.xsd")
	if err != nil {
		log.Fatalf("failed to parse taxonomy: %v", err)
	}
	doc.SetTaxonomy(tax)

	for _, f := range doc.Facts() {
		if f == nil {
			continue
		}
		c, ok := doc.ConceptOf(f)
		if !ok || c == nil {
			fmt.Printf("%s: concept not found in taxonomy\n", f.Name().String())
			continue
		}
		fmt.Printf("%s:\n", f.Name().String())
		fmt.Printf("  id          = %s\n", c.ID())
		fmt.Printf("  type        = %s\n", c.Type().String())
		fmt.Printf("  substGroup  = %s\n", c.SubstitutionGroup().String())
		fmt.Printf("  abstract    = %v\n", c.Abstract())
		fmt.Printf("  nillable    = %v\n", c.Nillable())
		fmt.Printf("  periodType  = %s\n", c.PeriodType())
		fmt.Printf("  balance     = %s\n", c.Balance())
		fmt.Println()
	}

	// --- Typed values based on Concept type ---
	fmt.Println("== Typed values based on Concept type ==")

	for _, f := range doc.Facts() {
		if f == nil {
			continue
		}
		c, ok := doc.ConceptOf(f)
		if !ok || c == nil {
			fmt.Printf("%s: concept not found, treat as raw string: %q\n",
				f.Name().String(), f.Value())
			continue
		}

		kind := c.ValueKind()
		fmt.Printf("%s:\n", f.Name().String())
		fmt.Printf("  valueKind = %s\n", kind)
		fmt.Printf("  raw       = %q\n", f.Value())

		switch kind {
		case xbrl.ConceptValueMonetary, xbrl.ConceptValueNumeric:
			if v, err := doc.AsInt64(f); err == nil {
				fmt.Printf("  AsInt64   = %d\n", v)
			} else {
				fmt.Printf("  AsInt64   error: %v\n", err)
			}
			if v, err := doc.AsFloat64(f); err == nil {
				fmt.Printf("  AsFloat64 = %f\n", v)
			} else {
				fmt.Printf("  AsFloat64 error: %v\n", err)
			}

		case xbrl.ConceptValueBoolean:
			if v, err := doc.AsBool(f); err == nil {
				fmt.Printf("  AsBool    = %v\n", v)
			} else {
				fmt.Printf("  AsBool    error: %v\n", err)
			}

		case xbrl.ConceptValueDate, xbrl.ConceptValueDateTime:
			if t, err := doc.AsTime(f, time.Local); err == nil {
				fmt.Printf("  AsTime    = %s\n", t.Format(time.RFC3339))
			} else {
				fmt.Printf("  AsTime    error: %v\n", err)
			}

		default:
			// treat as string
			fmt.Printf("  as string = %s\n", f.Value())
			if norm := f.NormalizedValue(); norm != f.Value() {
				fmt.Printf("  normalized= %s\n", norm)
			}
		}

		fmt.Println()
	}
}
