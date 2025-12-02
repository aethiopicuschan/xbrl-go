package xbrl_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/aethiopicuschan/xbrl-go/pkg/xbrl"
	"github.com/stretchr/testify/assert"
)

func TestSchemaRef_Href(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		sr   xbrl.SchemaRef
		want string
	}{
		{
			name: "non empty href",
			sr:   xbrl.NewSchemaRefForTest("file.xsd"),
			want: "file.xsd",
		},
		{
			name: "empty href",
			sr:   xbrl.NewSchemaRefForTest(""),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.sr.Href())
		})
	}
}

func TestContext_Basics(t *testing.T) {
	t.Parallel()

	ci := xbrl.NewContextIdentifierForTest("scheme", "value")
	entity := xbrl.NewEntityForTest(ci)
	var period xbrl.Period
	dim := xbrl.NewDimensionForTest(
		xbrl.NewQNameForTest("p", "d1", "u"),
		true,
		xbrl.NewQNameForTest("p", "m1", "u"),
		"",
	)
	ctx := xbrl.NewContextForTest("C1", entity, period, []xbrl.Dimension{dim})

	tests := []struct {
		name   string
		ctx    *xbrl.Context
		idWant string
		ent    xbrl.Entity
		per    xbrl.Period
	}{
		{
			name:   "non nil context",
			ctx:    ctx,
			idWant: "C1",
			ent:    entity,
			per:    period,
		},
		{
			name:   "nil context",
			ctx:    nil,
			idWant: "",
			ent:    xbrl.Entity{},
			per:    xbrl.Period{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.idWant, tt.ctx.ID())
			assert.Equal(t, tt.ent, tt.ctx.Entity())
			assert.Equal(t, tt.per, tt.ctx.Period())
		})
	}
}

func TestContext_DimensionsCopyAndLookup(t *testing.T) {
	t.Parallel()

	targetQName := xbrl.NewQNameForTest("p1", "dim", "uri")
	diffPrefixQName := xbrl.NewQNameForTest("p2", "dim", "uri") // same URI/local, different prefix
	nonMatchQName := xbrl.NewQNameForTest("p1", "dim2", "uri")

	dim1 := xbrl.NewDimensionForTest(targetQName, true, xbrl.NewQNameForTest("p", "m", "u"), "")
	dim2 := xbrl.NewDimensionForTest(nonMatchQName, false, xbrl.NewQNameForTest("", "", ""), "<typed/>")

	var emptyEntity xbrl.Entity
	var emptyPeriod xbrl.Period
	ctx := xbrl.NewContextForTest("C", emptyEntity, emptyPeriod, []xbrl.Dimension{dim1, dim2})

	t.Run("Dimensions returns copy", func(t *testing.T) {
		t.Parallel()

		got := ctx.Dimensions()
		assert.Equal(t, ctx.Dimensions(), got)

		// Modify returned slice and ensure original is not affected.
		got[0] = xbrl.Dimension{}
		again := ctx.Dimensions()
		assert.Equal(t, dim1, again[0])
	})

	t.Run("DimensionByQName matches by URI and local only", func(t *testing.T) {
		t.Parallel()

		got, ok := ctx.DimensionByQName(diffPrefixQName)
		assert.True(t, ok)
		assert.Equal(t, dim1, got)

		got, ok = ctx.DimensionByQName(nonMatchQName)
		assert.True(t, ok)
		assert.Equal(t, dim2, got)
	})

	t.Run("DimensionByQName nil context", func(t *testing.T) {
		t.Parallel()

		var nilCtx *xbrl.Context
		got, ok := nilCtx.DimensionByQName(targetQName)
		assert.False(t, ok)
		assert.Equal(t, xbrl.Dimension{}, got)
	})
}

func TestDimension_Methods(t *testing.T) {
	t.Parallel()

	q := xbrl.NewQNameForTest("p", "local", "uri")
	member := xbrl.NewQNameForTest("p2", "member", "uri2")

	tests := []struct {
		name       string
		d          xbrl.Dimension
		dimWant    xbrl.QName
		explicit   bool
		memberWant xbrl.QName
		typedWant  string
	}{
		{
			name:       "explicit dimension",
			d:          xbrl.NewDimensionForTest(q, true, member, ""),
			dimWant:    q,
			explicit:   true,
			memberWant: member,
			typedWant:  "",
		},
		{
			name:       "typed dimension",
			d:          xbrl.NewDimensionForTest(q, false, xbrl.NewQNameForTest("", "", ""), "<typed/>"),
			dimWant:    q,
			explicit:   false,
			memberWant: xbrl.NewQNameForTest("", "", ""),
			typedWant:  "<typed/>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.dimWant, tt.d.Dimension())
			assert.Equal(t, tt.explicit, tt.d.IsExplicit())
			assert.Equal(t, tt.memberWant, tt.d.Member())
			assert.Equal(t, tt.typedWant, tt.d.TypedValue())
		})
	}
}

func TestEntityAndContextIdentifier(t *testing.T) {
	t.Parallel()

	ci := xbrl.NewContextIdentifierForTest("scheme", "value")
	e := xbrl.NewEntityForTest(ci)

	t.Run("Entity Identifier", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, ci, e.Identifier())
	})

	t.Run("ContextIdentifier Scheme and Value", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "scheme", ci.Scheme())
		assert.Equal(t, "value", ci.Value())
	})
}

func TestPeriod_Methods(t *testing.T) {
	t.Parallel()

	inst := "2024-01-01"
	start := "2024-01-01"
	end := "2024-12-31"

	tests := []struct {
		name         string
		p            xbrl.Period
		instantOK    bool
		startOK      bool
		endOK        bool
		isInstant    bool
		isForever    bool
		instantValue string
		startValue   string
		endValue     string
	}{
		{
			name:         "instant only",
			p:            xbrl.NewPeriodForTest(&inst, nil, nil, false),
			instantOK:    true,
			startOK:      false,
			endOK:        false,
			isInstant:    true,
			isForever:    false,
			instantValue: inst,
			startValue:   "",
			endValue:     "",
		},
		{
			name:         "duration",
			p:            xbrl.NewPeriodForTest(nil, &start, &end, false),
			instantOK:    false,
			startOK:      true,
			endOK:        true,
			isInstant:    false,
			isForever:    false,
			instantValue: "",
			startValue:   start,
			endValue:     end,
		},
		{
			name:         "forever only",
			p:            xbrl.NewPeriodForTest(nil, nil, nil, true),
			instantOK:    false,
			startOK:      false,
			endOK:        false,
			isInstant:    false,
			isForever:    true,
			instantValue: "",
			startValue:   "",
			endValue:     "",
		},
		{
			name:         "empty period",
			p:            xbrl.NewPeriodForTest(nil, nil, nil, false),
			instantOK:    false,
			startOK:      false,
			endOK:        false,
			isInstant:    false,
			isForever:    false,
			instantValue: "",
			startValue:   "",
			endValue:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			instVal, ok := tt.p.Instant()
			assert.Equal(t, tt.instantOK, ok)
			assert.Equal(t, tt.instantValue, instVal)

			startVal, ok := tt.p.StartDate()
			assert.Equal(t, tt.startOK, ok)
			assert.Equal(t, tt.startValue, startVal)

			endVal, ok := tt.p.EndDate()
			assert.Equal(t, tt.endOK, ok)
			assert.Equal(t, tt.endValue, endVal)

			assert.Equal(t, tt.isInstant, tt.p.IsInstant())
			assert.Equal(t, tt.isForever, tt.p.IsForever())
		})
	}
}

func TestUnit_Methods(t *testing.T) {
	t.Parallel()

	m1 := xbrl.NewQNameForTest("p", "m1", "u")
	m2 := xbrl.NewQNameForTest("p", "m2", "u")

	unitSimple := xbrl.NewUnitSimpleForTest("U1", []xbrl.QName{m1, m2})
	unitDivide := xbrl.NewUnitDivideForTest("U2", []xbrl.QName{m1}, []xbrl.QName{m2})
	var nilUnit *xbrl.Unit

	t.Run("ID and nil", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "U1", unitSimple.ID())
		assert.Equal(t, "U2", unitDivide.ID())
		assert.Equal(t, "", nilUnit.ID())
	})

	t.Run("Measures returns copy", func(t *testing.T) {
		t.Parallel()

		got := unitSimple.Measures()
		assert.Equal(t, []xbrl.QName{m1, m2}, got)

		got[0] = xbrl.QName{}
		again := unitSimple.Measures()
		assert.Equal(t, m1, again[0])
	})

	t.Run("IsDivide and nil", func(t *testing.T) {
		t.Parallel()

		assert.False(t, unitSimple.IsDivide())
		assert.True(t, unitDivide.IsDivide())
		assert.False(t, nilUnit.IsDivide())
	})

	t.Run("Numerator and Denominator copies", func(t *testing.T) {
		t.Parallel()

		num := unitDivide.NumeratorMeasures()
		den := unitDivide.DenominatorMeasures()

		assert.Equal(t, []xbrl.QName{m1}, num)
		assert.Equal(t, []xbrl.QName{m2}, den)

		num[0] = xbrl.QName{}
		den[0] = xbrl.QName{}

		againNum := unitDivide.NumeratorMeasures()
		againDen := unitDivide.DenominatorMeasures()

		assert.Equal(t, m1, againNum[0])
		assert.Equal(t, m2, againDen[0])
	})

	t.Run("Nil unit collections", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, nilUnit.Measures())
		assert.Nil(t, nilUnit.NumeratorMeasures())
		assert.Nil(t, nilUnit.DenominatorMeasures())
	})
}

func TestQName_MethodsAndString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		q     xbrl.QName
		str   string
		p     string
		local string
		uri   string
	}{
		{
			name:  "empty",
			q:     xbrl.NewQNameForTest("", "", ""),
			str:   "",
			p:     "",
			local: "",
			uri:   "",
		},
		{
			name:  "no uri with prefix",
			q:     xbrl.NewQNameForTest("p", "local", ""),
			str:   "p:local",
			p:     "p",
			local: "local",
			uri:   "",
		},
		{
			name:  "no uri no prefix",
			q:     xbrl.NewQNameForTest("", "local", ""),
			str:   "local",
			p:     "",
			local: "local",
			uri:   "",
		},
		{
			name:  "with uri",
			q:     xbrl.NewQNameForTest("p", "local", "urn:ns"),
			str:   "{urn:ns}local",
			p:     "p",
			local: "local",
			uri:   "urn:ns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.p, tt.q.Prefix())
			assert.Equal(t, tt.local, tt.q.Local())
			assert.Equal(t, tt.uri, tt.q.URI())
			assert.Equal(t, tt.str, tt.q.String())
		})
	}
}

func TestFact_Methods(t *testing.T) {
	t.Parallel()

	name := xbrl.NewQNameForTest("p", "local", "uri")
	f := xbrl.NewFactForTest(
		xbrl.FactKindItem,
		name,
		"  foo\tbar\nbaz  ",
		"C1",
		"U1",
		"2",
		"3",
		"F1",
		"en",
		true,
	)
	var nilFact *xbrl.Fact

	t.Run("Kind and nil", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, xbrl.FactKindItem, f.Kind())
		assert.Equal(t, xbrl.FactKindUnknown, nilFact.Kind())
	})

	t.Run("Name and Value", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, name, f.Name())
		assert.Equal(t, xbrl.QName{}, nilFact.Name())
		assert.Equal(t, "  foo\tbar\nbaz  ", f.Value())
		assert.Equal(t, "", nilFact.Value())
	})

	t.Run("NormalizedValue", func(t *testing.T) {
		t.Parallel()

		// Whitespace should be normalized.
		norm := f.NormalizedValue()
		assert.Equal(t, "foo bar baz", norm)
		assert.Equal(t, "", nilFact.NormalizedValue())
	})

	t.Run("References and attributes", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "C1", f.ContextRef())
		assert.Equal(t, "U1", f.UnitRef())
		assert.Equal(t, "2", f.Decimals())
		assert.Equal(t, "3", f.Precision())
		assert.Equal(t, "F1", f.ID())
		assert.Equal(t, "en", f.Lang())
		assert.Equal(t, "", nilFact.ContextRef())
		assert.Equal(t, "", nilFact.UnitRef())
		assert.Equal(t, "", nilFact.Decimals())
		assert.Equal(t, "", nilFact.Precision())
		assert.Equal(t, "", nilFact.ID())
		assert.Equal(t, "", nilFact.Lang())
	})

	t.Run("IsNil", func(t *testing.T) {
		t.Parallel()

		assert.True(t, f.IsNil())
		assert.False(t, nilFact.IsNil())
	})
}

func TestConcept_MethodsAndKinds(t *testing.T) {
	t.Parallel()

	itemSG := xbrl.NewQNameForTest("xbrli", "item", "http://www.xbrl.org/2003/instance")
	tupleSG := xbrl.NewQNameForTest("xbrli", "tuple", "http://www.xbrl.org/2003/instance")
	otherSG := xbrl.NewQNameForTest("x", "y", "urn:other")

	q := xbrl.NewQNameForTest("p", "c", "uri")
	typ := xbrl.NewQNameForTest("xbrli", "monetaryItemType", "http://www.xbrl.org/2003/instance")

	emptyQName := xbrl.NewQNameForTest("", "", "")

	conceptItem := xbrl.NewConceptForTest(q, "C1", itemSG, typ, true, true, "duration", "debit")
	conceptTuple := xbrl.NewConceptForTest(q, "", tupleSG, emptyQName, false, false, "", "")
	conceptOther := xbrl.NewConceptForTest(q, "", otherSG, emptyQName, false, false, "", "")
	var nilConcept *xbrl.Concept

	t.Run("Basic getters and nil", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, q, conceptItem.QName())
		assert.Equal(t, "C1", conceptItem.ID())
		assert.Equal(t, itemSG, conceptItem.SubstitutionGroup())
		assert.Equal(t, typ, conceptItem.Type())
		assert.True(t, conceptItem.Abstract())
		assert.True(t, conceptItem.Nillable())
		assert.Equal(t, "duration", conceptItem.PeriodType())
		assert.Equal(t, "debit", conceptItem.Balance())

		assert.Equal(t, xbrl.QName{}, nilConcept.QName())
		assert.Equal(t, "", nilConcept.ID())
		assert.Equal(t, xbrl.QName{}, nilConcept.SubstitutionGroup())
		assert.Equal(t, xbrl.QName{}, nilConcept.Type())
		assert.False(t, nilConcept.Abstract())
		assert.False(t, nilConcept.Nillable())
		assert.Equal(t, "", nilConcept.PeriodType())
		assert.Equal(t, "", nilConcept.Balance())
	})

	t.Run("IsItem and IsTuple", func(t *testing.T) {
		t.Parallel()

		assert.True(t, conceptItem.IsItem())
		assert.False(t, conceptItem.IsTuple())

		assert.False(t, conceptTuple.IsItem())
		assert.True(t, conceptTuple.IsTuple())

		assert.False(t, conceptOther.IsItem())
		assert.False(t, conceptOther.IsTuple())

		assert.False(t, nilConcept.IsItem())
		assert.False(t, nilConcept.IsTuple())
	})
}

func TestTaxonomy_Methods(t *testing.T) {
	t.Parallel()

	q1 := xbrl.NewQNameForTest("p", "c1", "u")
	q2 := xbrl.NewQNameForTest("p", "c2", "u")

	emptyQName := xbrl.NewQNameForTest("", "", "")
	c1 := xbrl.NewConceptForTest(q1, "", emptyQName, emptyQName, false, false, "", "")
	c2 := xbrl.NewConceptForTest(q2, "", emptyQName, emptyQName, false, false, "", "")

	t.Run("NewTaxonomy and empty", func(t *testing.T) {
		t.Parallel()

		tax := xbrl.NewTaxonomy()
		assert.NotNil(t, tax)

		// Concepts map should be an empty copy.
		concepts := tax.Concepts()
		assert.Empty(t, concepts)

		// Looking up concept should fail.
		got, ok := tax.Concept(q1)
		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("Concepts returns copy", func(t *testing.T) {
		t.Parallel()

		tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{
			q1: c1,
			q2: c2,
		})

		cmap := tax.Concepts()
		assert.Len(t, cmap, 2)

		// Modifying returned map should not affect internal map.
		delete(cmap, q1)
		again := tax.Concepts()
		assert.Len(t, again, 2)
	})

	t.Run("Concept lookup and nil taxonomy", func(t *testing.T) {
		t.Parallel()

		tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{
			q1: c1,
		})

		got, ok := tax.Concept(q1)
		assert.True(t, ok)
		assert.Equal(t, c1, got)

		gotNil, okNil := tax.Concept(q2)
		assert.False(t, okNil)
		assert.Nil(t, gotNil)

		var nilTax *xbrl.Taxonomy
		gotNil, okNil = nilTax.Concept(q1)
		assert.False(t, okNil)
		assert.Nil(t, gotNil)

		assert.Nil(t, nilTax.Concepts())
	})
}

func TestDocument_CollectionsAndLookup(t *testing.T) {
	t.Parallel()

	var emptyEntity xbrl.Entity
	var emptyPeriod xbrl.Period
	ctx := xbrl.NewContextForTest("C1", emptyEntity, emptyPeriod, nil)
	unit := xbrl.NewUnitSimpleForTest("U1", nil)
	emptyQName := xbrl.NewQNameForTest("", "", "")
	fact := xbrl.NewFactForTest(
		xbrl.FactKindUnknown,
		emptyQName,
		"",
		"C1",
		"U1",
		"",
		"",
		"",
		"",
		false,
	)

	doc := xbrl.NewDocumentForTest(
		[]xbrl.SchemaRef{
			xbrl.NewSchemaRefForTest("a.xsd"),
			xbrl.NewSchemaRefForTest("b.xsd"),
		},
		map[string]*xbrl.Context{"C1": ctx},
		map[string]*xbrl.Unit{"U1": unit},
		[]*xbrl.Fact{fact},
		nil,
	)
	var nilDoc *xbrl.Document

	t.Run("SchemaRefs returns copy and nil", func(t *testing.T) {
		t.Parallel()

		srs := doc.SchemaRefs()
		assert.Equal(t, 2, len(srs))
		srs[0] = xbrl.NewSchemaRefForTest("modified.xsd")
		again := doc.SchemaRefs()
		assert.Equal(t, "a.xsd", again[0].Href())

		assert.Nil(t, nilDoc.SchemaRefs())
	})

	t.Run("Contexts returns copy and nil", func(t *testing.T) {
		t.Parallel()

		cmap := doc.Contexts()
		assert.Len(t, cmap, 1)
		delete(cmap, "C1")
		again := doc.Contexts()
		assert.Len(t, again, 1)

		assert.Nil(t, nilDoc.Contexts())
	})

	t.Run("Units returns copy and nil", func(t *testing.T) {
		t.Parallel()

		umap := doc.Units()
		assert.Len(t, umap, 1)
		delete(umap, "U1")
		again := doc.Units()
		assert.Len(t, again, 1)

		assert.Nil(t, nilDoc.Units())
	})

	t.Run("Facts returns copy and nil", func(t *testing.T) {
		t.Parallel()

		fs := doc.Facts()
		assert.Len(t, fs, 1)
		fs[0] = xbrl.NewFactForTest(
			xbrl.FactKindUnknown,
			emptyQName,
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			false,
		)
		again := doc.Facts()
		assert.Equal(t, fact, again[0])

		assert.Nil(t, nilDoc.Facts())
	})

	t.Run("ContextByID and UnitByID with nil document", func(t *testing.T) {
		t.Parallel()

		gotCtx, ok := doc.ContextByID("C1")
		assert.True(t, ok)
		assert.Equal(t, ctx, gotCtx)

		_, ok = doc.ContextByID("missing")
		assert.False(t, ok)

		gotUnit, ok := doc.UnitByID("U1")
		assert.True(t, ok)
		assert.Equal(t, unit, gotUnit)

		_, ok = doc.UnitByID("missing")
		assert.False(t, ok)

		gotCtx, ok = nilDoc.ContextByID("C1")
		assert.False(t, ok)
		assert.Nil(t, gotCtx)

		gotUnit, ok = nilDoc.UnitByID("U1")
		assert.False(t, ok)
		assert.Nil(t, gotUnit)
	})

	t.Run("ContextOf and UnitOf", func(t *testing.T) {
		t.Parallel()

		gotCtx, ok := doc.ContextOf(fact)
		assert.True(t, ok)
		assert.Equal(t, ctx, gotCtx)

		gotUnit, ok := doc.UnitOf(fact)
		assert.True(t, ok)
		assert.Equal(t, unit, gotUnit)

		// Nil document or nil fact should return false.
		gotCtx, ok = nilDoc.ContextOf(fact)
		assert.False(t, ok)
		assert.Nil(t, gotCtx)

		gotUnit, ok = nilDoc.UnitOf(fact)
		assert.False(t, ok)
		assert.Nil(t, gotUnit)

		gotCtx, ok = doc.ContextOf(nil)
		assert.False(t, ok)
		assert.Nil(t, gotCtx)

		gotUnit, ok = doc.UnitOf(nil)
		assert.False(t, ok)
		assert.Nil(t, gotUnit)
	})
}

func TestDocument_TaxonomyAndConceptOf(t *testing.T) {
	t.Parallel()

	q := xbrl.NewQNameForTest("p", "c1", "u")
	emptyQName := xbrl.NewQNameForTest("", "", "")
	concept := xbrl.NewConceptForTest(q, "", emptyQName, emptyQName, false, false, "", "")
	tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{
		q: concept,
	})

	fact := xbrl.NewFactForTest(
		xbrl.FactKindItem,
		q,
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		false,
	)
	doc := xbrl.NewDocumentForTest(nil, nil, nil, nil, tax)
	var nilDoc *xbrl.Document

	t.Run("Taxonomy getter and setter", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, tax, doc.Taxonomy())

		doc2 := new(xbrl.Document)
		assert.Nil(t, doc2.Taxonomy())
		doc2.SetTaxonomy(tax)
		assert.Equal(t, tax, doc2.Taxonomy())

		// Nil document should do nothing.
		nilDoc.SetTaxonomy(tax)
		assert.Nil(t, nilDoc.Taxonomy())
	})

	t.Run("ConceptOf success and misses", func(t *testing.T) {
		t.Parallel()

		got, ok := doc.ConceptOf(fact)
		assert.True(t, ok)
		assert.Equal(t, concept, got)

		// Missing concept
		otherFact := xbrl.NewFactForTest(
			xbrl.FactKindItem,
			xbrl.NewQNameForTest("p", "c2", "u"),
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			false,
		)
		got, ok = doc.ConceptOf(otherFact)
		assert.False(t, ok)
		assert.Nil(t, got)

		// Nil document or fact
		got, ok = nilDoc.ConceptOf(fact)
		assert.False(t, ok)
		assert.Nil(t, got)

		got, ok = doc.ConceptOf(nil)
		assert.False(t, ok)
		assert.Nil(t, got)

		// Document without taxonomy
		docNoTax := new(xbrl.Document)
		got, ok = docNoTax.ConceptOf(fact)
		assert.False(t, ok)
		assert.Nil(t, got)
	})
}

func TestDocument_LoadTaxonomyFromSchemaRefs_ErrorsAndBasics(t *testing.T) {
	t.Parallel()

	docWithNil := (*xbrl.Document)(nil)

	t.Run("nil document", func(t *testing.T) {
		t.Parallel()

		tax, err := docWithNil.LoadTaxonomyFromSchemaRefs(func(string) (io.ReadCloser, error) {
			return nil, nil
		})
		assert.Nil(t, tax)
		assert.EqualError(t, err, "xbrl: document is nil")
	})

	doc := xbrl.NewDocumentForTest(
		[]xbrl.SchemaRef{
			xbrl.NewSchemaRefForTest(""),
			xbrl.NewSchemaRefForTest("a.xsd"),
			xbrl.NewSchemaRefForTest("broken.xsd"),
		},
		nil,
		nil,
		nil,
		nil,
	)

	t.Run("nil opener", func(t *testing.T) {
		t.Parallel()

		tax, err := doc.LoadTaxonomyFromSchemaRefs(nil)
		assert.Nil(t, tax)
		assert.EqualError(t, err, "xbrl: opener is nil")
	})

	t.Run("opener error", func(t *testing.T) {
		t.Parallel()

		opener := func(href string) (io.ReadCloser, error) {
			if href == "a.xsd" {
				return io.NopCloser(strings.NewReader("")), nil
			}
			return nil, errors.New("open failed")
		}

		tax, err := doc.LoadTaxonomyFromSchemaRefs(opener)
		assert.Nil(t, tax)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `xbrl: open schemaRef "broken.xsd"`)
	})

	t.Run("success with empty schema", func(t *testing.T) {
		t.Parallel()

		docOK := xbrl.NewDocumentForTest(
			[]xbrl.SchemaRef{
				xbrl.NewSchemaRefForTest(""),
				xbrl.NewSchemaRefForTest("a.xsd"),
			},
			nil,
			nil,
			nil,
			nil,
		)

		opener := func(href string) (io.ReadCloser, error) {
			// Provide empty content; ParseTaxonomy will succeed with empty taxonomy.
			return io.NopCloser(strings.NewReader("")), nil
		}

		tax, err := docOK.LoadTaxonomyFromSchemaRefs(opener)
		assert.NoError(t, err)
		assert.NotNil(t, tax)
		assert.NotNil(t, docOK.Taxonomy())
		assert.Same(t, tax, docOK.Taxonomy())
	})
}
