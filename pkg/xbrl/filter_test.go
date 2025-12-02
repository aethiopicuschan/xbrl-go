package xbrl_test

import (
	"testing"

	"github.com/aethiopicuschan/xbrl-go/pkg/xbrl"
	"github.com/stretchr/testify/assert"
)

// Test that builder-style methods are safe on nil receiver and return nil.
func TestFactFilter_NilReceiver(t *testing.T) {
	t.Parallel()

	var f *xbrl.FactFilter
	dim := xbrl.NewQNameForTest("d", "dim", "urn:dim")
	mem := xbrl.NewQNameForTest("m", "mem", "urn:mem")

	tests := []struct {
		name string
		call func() *xbrl.FactFilter
	}{
		{
			name: "ConceptURI on nil",
			call: func() *xbrl.FactFilter { return f.ConceptURI("uri") },
		},
		{
			name: "ConceptLocal on nil",
			call: func() *xbrl.FactFilter { return f.ConceptLocal("local") },
		},
		{
			name: "ContextID on nil",
			call: func() *xbrl.FactFilter { return f.ContextID("ctx") },
		},
		{
			name: "UnitID on nil",
			call: func() *xbrl.FactFilter { return f.UnitID("unit") },
		},
		{
			name: "OnlyNil on nil",
			call: func() *xbrl.FactFilter { return f.OnlyNil() },
		},
		{
			name: "ExcludeNil on nil",
			call: func() *xbrl.FactFilter { return f.ExcludeNil() },
		},
		{
			name: "Dimension on nil",
			call: func() *xbrl.FactFilter { return f.Dimension(dim, mem) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.call()
			assert.Nil(t, got, "method on nil receiver should return nil")
		})
	}
}

// Test that builder-style methods can be chained and affect filtering behavior.
func TestFactFilter_BuilderAndFilteringBasics(t *testing.T) {
	t.Parallel()

	// Prepare three facts with different concept/context/unit/nil combinations.
	q1 := xbrl.NewQNameForTest("p", "x", "urn:a")
	q2 := xbrl.NewQNameForTest("p", "y", "urn:b")

	f1 := xbrl.NewFactForTest(
		xbrl.FactKindItem, // kind
		q1,                // name
		"v1",              // value
		"C1",              // contextRef
		"U1",              // unitRef
		"",                // decimals
		"",                // precision
		"F1",              // id
		"",                // lang
		false,             // nil
	)
	f2 := xbrl.NewFactForTest(
		xbrl.FactKindItem,
		q2,
		"v2",
		"C1",
		"U1",
		"",
		"",
		"F2",
		"",
		true, // nil = true
	)
	f3 := xbrl.NewFactForTest(
		xbrl.FactKindItem,
		q1,
		"v3",
		"C2",
		"U2",
		"",
		"",
		"F3",
		"",
		false,
	)

	doc := xbrl.NewDocumentForTest(
		nil,
		nil,
		nil,
		[]*xbrl.Fact{f1, f2, f3},
		nil,
	)

	tests := []struct {
		name   string
		filter *xbrl.FactFilter
		want   []*xbrl.Fact
	}{
		{
			name:   "empty filter returns all facts",
			filter: xbrl.NewFactFilter(),
			want:   []*xbrl.Fact{f1, f2, f3},
		},
		{
			name:   "concept local only",
			filter: xbrl.NewFactFilter().ConceptLocal("x"),
			want:   []*xbrl.Fact{f1, f3},
		},
		{
			name:   "concept URI only",
			filter: xbrl.NewFactFilter().ConceptURI("urn:a"),
			want:   []*xbrl.Fact{f1, f3},
		},
		{
			name:   "concept URI and local",
			filter: xbrl.NewFactFilter().ConceptURI("urn:a").ConceptLocal("x"),
			want:   []*xbrl.Fact{f1, f3},
		},
		{
			name:   "context ID",
			filter: xbrl.NewFactFilter().ContextID("C1"),
			want:   []*xbrl.Fact{f1, f2},
		},
		{
			name:   "unit ID",
			filter: xbrl.NewFactFilter().UnitID("U2"),
			want:   []*xbrl.Fact{f3},
		},
		{
			name:   "combined concept and context",
			filter: xbrl.NewFactFilter().ConceptLocal("x").ContextID("C2"),
			want:   []*xbrl.Fact{f3},
		},
		{
			name:   "OnlyNil keeps only nil facts",
			filter: xbrl.NewFactFilter().OnlyNil(),
			want:   []*xbrl.Fact{f2},
		},
		{
			name:   "ExcludeNil keeps only non-nil facts",
			filter: xbrl.NewFactFilter().ExcludeNil(),
			want:   []*xbrl.Fact{f1, f3},
		},
		{
			name:   "OnlyNil overrides ExcludeNil when chained last",
			filter: xbrl.NewFactFilter().ExcludeNil().OnlyNil(),
			want:   []*xbrl.Fact{f2},
		},
		{
			name:   "ExcludeNil overrides OnlyNil when chained last",
			filter: xbrl.NewFactFilter().OnlyNil().ExcludeNil(),
			want:   []*xbrl.Fact{f1, f3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := doc.FilterFacts(tt.filter)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Test that FilterFacts handles nil document and nil filter safely.
func TestDocument_FilterFacts_NilDocOrFilter(t *testing.T) {
	t.Parallel()

	var nilDoc *xbrl.Document
	doc := xbrl.NewDocumentForTest(nil, nil, nil, nil, nil)
	filter := xbrl.NewFactFilter()

	tests := []struct {
		name   string
		doc    *xbrl.Document
		filter *xbrl.FactFilter
	}{
		{
			name:   "nil document",
			doc:    nilDoc,
			filter: filter,
		},
		{
			name:   "nil filter",
			doc:    doc,
			filter: nil,
		},
		{
			name:   "nil doc and nil filter",
			doc:    nilDoc,
			filter: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.doc.FilterFacts(tt.filter)
			assert.Nil(t, got)
		})
	}
}

// Test that FilterFacts applies dimension filters correctly and only considers explicit dimensions.
func TestDocument_FilterFacts_Dimensions(t *testing.T) {
	t.Parallel()

	// Common concept and unit.
	conceptQName := xbrl.NewQNameForTest("c", "item", "urn:concept")
	unitID := "U1"

	// Dimension and members.
	dimQName := xbrl.NewQNameForTest("d", "dim1", "urn:dim")
	mem1 := xbrl.NewQNameForTest("m", "mem1", "urn:mem")
	mem2 := xbrl.NewQNameForTest("m", "mem2", "urn:mem")

	// Explicit dimensions.
	expDim1 := xbrl.NewDimensionForTest(dimQName, true, mem1, "")
	expDim2 := xbrl.NewDimensionForTest(dimQName, true, mem2, "")

	// Typed dimension with same dim but no explicit member.
	typedDim := xbrl.NewDimensionForTest(dimQName, false, xbrl.NewQNameForTest("", "", ""), "<typed/>")

	var emptyEntity xbrl.Entity
	var emptyPeriod xbrl.Period

	// Contexts:
	// C1: explicit mem1 + typed dim
	ctx1 := xbrl.NewContextForTest("C1", emptyEntity, emptyPeriod, []xbrl.Dimension{expDim1, typedDim})
	// C2: explicit mem2 only
	ctx2 := xbrl.NewContextForTest("C2", emptyEntity, emptyPeriod, []xbrl.Dimension{expDim2})
	// C3: typed dim only (no explicit)
	ctx3 := xbrl.NewContextForTest("C3", emptyEntity, emptyPeriod, []xbrl.Dimension{typedDim})
	// C4: explicit mem1 and mem2
	ctx4 := xbrl.NewContextForTest("C4", emptyEntity, emptyPeriod, []xbrl.Dimension{expDim1, expDim2})

	// Facts referring to these contexts.
	f1 := xbrl.NewFactForTest(
		xbrl.FactKindItem, conceptQName, "v1", "C1", unitID, "", "", "F1", "", false,
	)
	f2 := xbrl.NewFactForTest(
		xbrl.FactKindItem, conceptQName, "v2", "C2", unitID, "", "", "F2", "", false,
	)
	f3 := xbrl.NewFactForTest(
		xbrl.FactKindItem, conceptQName, "v3", "C3", unitID, "", "", "F3", "", false,
	)
	f4 := xbrl.NewFactForTest(
		xbrl.FactKindItem, conceptQName, "v4", "C4", unitID, "", "", "F4", "", false,
	)
	// Fact with context that is missing from document.
	fMissingCtx := xbrl.NewFactForTest(
		xbrl.FactKindItem, conceptQName, "v5", "MISSING", unitID, "", "", "F5", "", false,
	)

	doc := xbrl.NewDocumentForTest(
		nil,
		map[string]*xbrl.Context{
			"C1": ctx1,
			"C2": ctx2,
			"C3": ctx3,
			"C4": ctx4,
			// "MISSING" is intentionally not present.
		},
		nil,
		[]*xbrl.Fact{f1, f2, f3, f4, fMissingCtx},
		nil,
	)

	tests := []struct {
		name   string
		filter *xbrl.FactFilter
		want   []*xbrl.Fact
	}{
		{
			name:   "single dimension (dim1=mem1) matches C1 and C4",
			filter: xbrl.NewFactFilter().Dimension(dimQName, mem1),
			want:   []*xbrl.Fact{f1, f4},
		},
		{
			name:   "single dimension (dim1=mem2) matches C2 and C4",
			filter: xbrl.NewFactFilter().Dimension(dimQName, mem2),
			want:   []*xbrl.Fact{f2, f4},
		},
		{
			name:   "dimension with no matching explicit member yields empty result",
			filter: xbrl.NewFactFilter().Dimension(dimQName, xbrl.NewQNameForTest("m", "other", "urn:mem")),
			want:   []*xbrl.Fact{},
		},
		{
			name:   "typed-only context does not match explicit dimension requirement",
			filter: xbrl.NewFactFilter().Dimension(dimQName, mem1),
			// f3 (C3) is not included because it has only typed dimension.
			want: []*xbrl.Fact{f1, f4},
		},
		{
			name:   "multiple dimension requirements must all match (C4 only)",
			filter: xbrl.NewFactFilter().Dimension(dimQName, mem1).Dimension(dimQName, mem2),
			want:   []*xbrl.Fact{f4},
		},
		{
			name:   "fact with missing context is skipped",
			filter: xbrl.NewFactFilter().Dimension(dimQName, mem1),
			// fMissingCtx is not returned because its context is not found.
			want: []*xbrl.Fact{f1, f4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := doc.FilterFacts(tt.filter)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Test that FilterFacts returns a shallow copy slice (caller can modify it without
// affecting subsequent calls).
func TestDocument_FilterFacts_ReturnsCopy(t *testing.T) {
	t.Parallel()

	q := xbrl.NewQNameForTest("p", "x", "urn:a")
	f1 := xbrl.NewFactForTest(
		xbrl.FactKindItem, q, "v1", "C1", "U1", "", "", "F1", "", false,
	)
	f2 := xbrl.NewFactForTest(
		xbrl.FactKindItem, q, "v2", "C1", "U1", "", "", "F2", "", false,
	)

	doc := xbrl.NewDocumentForTest(nil, nil, nil, []*xbrl.Fact{f1, f2}, nil)
	filter := xbrl.NewFactFilter()

	first := doc.FilterFacts(filter)
	assert.Len(t, first, 2)

	// Modify returned slice and ensure internal data is not affected.
	first[0] = nil

	second := doc.FilterFacts(filter)
	assert.Len(t, second, 2)
	assert.Equal(t, f1, second[0])
	assert.Equal(t, f2, second[1])
}
