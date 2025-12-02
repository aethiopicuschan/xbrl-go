package xbrl_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aethiopicuschan/xbrl-go/pkg/xbrl"
	"github.com/stretchr/testify/assert"
)

// TestParseTaxonomy_EmptySchema verifies that an empty schema produces
// an empty taxonomy without error.
func TestParseTaxonomy_EmptySchema(t *testing.T) {
	t.Parallel()

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="http://example.com/tax"
           xmlns="http://example.com/tax">
</xs:schema>`

	tax, err := xbrl.ParseTaxonomy(strings.NewReader(xml))
	assert.NoError(t, err)
	assert.NotNil(t, tax)
	assert.Empty(t, tax.Concepts())
}

// TestParseTaxonomy_SingleElement verifies that a single xs:element is
// converted into a Concept with the expected fields.
func TestParseTaxonomy_SingleElement(t *testing.T) {
	t.Parallel()

	const targetNS = "http://example.com/tax"

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           xmlns:xbrli="http://www.xbrl.org/2003/instance"
           targetNamespace="` + targetNS + `"
           xmlns="` + targetNS + `">
  <xs:element name="Foo" id="Foo_1"
              substitutionGroup="xbrli:item"
              type="xbrli:monetaryItemType"
              abstract="true"
              nillable="1"
              periodType="duration"
              balance="debit"/>
</xs:schema>`

	tax, err := xbrl.ParseTaxonomy(strings.NewReader(xml))
	assert.NoError(t, err)
	assert.NotNil(t, tax)

	// Concept QName is (targetNS, "Foo").
	qFoo := xbrl.NewQNameForTest("", "Foo", targetNS)

	c, ok := tax.Concept(qFoo)
	if assert.True(t, ok, "concept Foo should exist") && assert.NotNil(t, c) {
		q := c.QName()
		assert.Equal(t, "Foo", q.Local())
		assert.Equal(t, targetNS, q.URI())

		sg := c.SubstitutionGroup()
		assert.Equal(t, "item", sg.Local())
		assert.Equal(t, "http://www.xbrl.org/2003/instance", sg.URI())

		typ := c.Type()
		assert.Equal(t, "monetaryItemType", typ.Local())
		assert.Equal(t, "http://www.xbrl.org/2003/instance", typ.URI())

		assert.True(t, c.Abstract())
		assert.True(t, c.Nillable())
		assert.Equal(t, "duration", c.PeriodType())
		assert.Equal(t, "debit", c.Balance())
	}
}

// TestParseTaxonomy_ParseBoolBehavior indirectly checks parseBool()
// via abstract/nillable attributes with various lexical forms.
func TestParseTaxonomy_ParseBoolBehavior(t *testing.T) {
	t.Parallel()

	const targetNS = "http://example.com/bool"

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="` + targetNS + `"
           xmlns="` + targetNS + `">
  <xs:element name="ETrue" abstract="true" nillable="true"/>
  <xs:element name="ETrueUpper" abstract="TRUE" nillable="1"/>
  <xs:element name="EOne" abstract="1" nillable="1"/>
  <xs:element name="EFalse" abstract="false" nillable="0"/>
  <xs:element name="EOther" abstract="yes" nillable="no"/>
</xs:schema>`

	tax, err := xbrl.ParseTaxonomy(strings.NewReader(xml))
	assert.NoError(t, err)
	assert.NotNil(t, tax)

	type want struct {
		abs bool
		nil bool
	}

	cases := []struct {
		name string
		want want
	}{
		{"ETrue", want{true, true}},
		{"ETrueUpper", want{true, true}},
		{"EOne", want{true, true}},
		{"EFalse", want{false, false}},
		{"EOther", want{false, false}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			q := xbrl.NewQNameForTest("", tt.name, targetNS)
			c, ok := tax.Concept(q)
			if assert.True(t, ok, "concept %s should exist", tt.name) && assert.NotNil(t, c) {
				assert.Equal(t, tt.want.abs, c.Abstract())
				assert.Equal(t, tt.want.nil, c.Nillable())
			}
		})
	}
}

// TestParseTaxonomy_MissingNameOrTargetNS ensures that elements without
// a name or schemas without targetNamespace do not produce concepts.
func TestParseTaxonomy_MissingNameOrTargetNS(t *testing.T) {
	t.Parallel()

	t.Run("missing targetNamespace", func(t *testing.T) {
		t.Parallel()

		xml := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="Foo" />
</xs:schema>`

		tax, err := xbrl.ParseTaxonomy(strings.NewReader(xml))
		assert.NoError(t, err)
		assert.NotNil(t, tax)
		assert.Empty(t, tax.Concepts(), "no targetNamespace -> no concepts")
	})

	t.Run("element with no name", func(t *testing.T) {
		t.Parallel()

		xml := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="http://example.com/tax"
           xmlns="http://example.com/tax">
  <xs:element id="FooId" />
</xs:schema>`

		tax, err := xbrl.ParseTaxonomy(strings.NewReader(xml))
		assert.NoError(t, err)
		assert.NotNil(t, tax)
		assert.Empty(t, tax.Concepts(), "element without name should be ignored")
	})
}

// TestParseTaxonomy_InvalidXML ensures that XML decoder errors are wrapped
// and returned properly.
func TestParseTaxonomy_InvalidXML(t *testing.T) {
	t.Parallel()

	// Malformed XML (unclosed element).
	xml := `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="Foo">
</xs:schema`

	tax, err := xbrl.ParseTaxonomy(strings.NewReader(xml))
	assert.Nil(t, tax)

	if assert.Error(t, err) {
		// Implementation may fail while reading tokens or while skipping an element.
		// We only assert that the error is wrapped with the "xbrl:" prefix.
		msg := err.Error()
		assert.Contains(t, msg, "xbrl:", "unexpected error message: %s", msg)
	}
}

// TestParseTaxonomy_MultipleElements ensures that multiple xs:element
// declarations produce multiple concepts in the taxonomy.
func TestParseTaxonomy_MultipleElements(t *testing.T) {
	t.Parallel()

	const targetNS = "http://example.com/multi"

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="` + targetNS + `"
           xmlns="` + targetNS + `">
  <xs:element name="A"/>
  <xs:element name="B"/>
</xs:schema>`

	tax, err := xbrl.ParseTaxonomy(strings.NewReader(xml))
	assert.NoError(t, err)
	assert.NotNil(t, tax)

	// We expect two concepts: A and B.
	qA := xbrl.NewQNameForTest("", "A", targetNS)
	qB := xbrl.NewQNameForTest("", "B", targetNS)

	cA, okA := tax.Concept(qA)
	cB, okB := tax.Concept(qB)

	assert.True(t, okA)
	assert.True(t, okB)
	assert.NotNil(t, cA)
	assert.NotNil(t, cB)

	concepts := tax.Concepts()
	assert.Len(t, concepts, 2)
}

// TestParseTaxonomyFile_SuccessAndOpenError covers ParseTaxonomyFile for
// both successful and error cases.
func TestParseTaxonomyFile_SuccessAndOpenError(t *testing.T) {
	t.Parallel()

	t.Run("open error", func(t *testing.T) {
		t.Parallel()

		// Use a path in a temp dir that we do not create.
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "does-not-exist.xsd")

		tax, err := xbrl.ParseTaxonomyFile(path)
		assert.Nil(t, tax)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "xbrl: open taxonomy schema")
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		const targetNS = "http://example.com/file"

		xml := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="` + targetNS + `"
           xmlns="` + targetNS + `">
  <xs:element name="FromFile"/>
</xs:schema>`

		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "tax.xsd")

		err := os.WriteFile(path, []byte(xml), 0o644)
		assert.NoError(t, err)

		tax, err := xbrl.ParseTaxonomyFile(path)
		assert.NoError(t, err)
		assert.NotNil(t, tax)

		q := xbrl.NewQNameForTest("", "FromFile", targetNS)
		c, ok := tax.Concept(q)
		assert.True(t, ok)
		assert.NotNil(t, c)
	})
}

// TestTaxonomy_Merge verifies that Merge correctly merges concept maps,
// handles nil arguments, and initializes a nil concepts map.
func TestTaxonomy_Merge(t *testing.T) {
	t.Parallel()

	q1 := xbrl.NewQNameForTest("p", "A", "urn:test")
	q2 := xbrl.NewQNameForTest("p", "B", "urn:test")

	emptyQName := xbrl.NewQNameForTest("", "", "")

	// c1 and c1Override share the same QName but differ in ID.
	c1 := xbrl.NewConceptForTest(q1, "C1", emptyQName, emptyQName, false, false, "", "")
	c1Override := xbrl.NewConceptForTest(q1, "C1-override", emptyQName, emptyQName, false, false, "", "")
	c2 := xbrl.NewConceptForTest(q2, "C2", emptyQName, emptyQName, false, false, "", "")

	t.Run("merge into taxonomy with initialized map", func(t *testing.T) {
		t.Parallel()

		left := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{
			q1: c1,
		})
		right := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{
			q1: c1Override, // should overwrite
			q2: c2,
		})

		left.Merge(right)

		concepts := left.Concepts()
		if assert.Len(t, concepts, 2) {
			got1, ok1 := left.Concept(q1)
			got2, ok2 := left.Concept(q2)

			assert.True(t, ok1)
			assert.True(t, ok2)

			// q1 should be overwritten by c1Override.
			assert.Equal(t, "C1-override", got1.ID())
			assert.Equal(t, "C2", got2.ID())
		}

		// Ensure right taxonomy is unchanged.
		rightConcepts := right.Concepts()
		assert.Len(t, rightConcepts, 2)
	})

	t.Run("merge into taxonomy with nil concepts map", func(t *testing.T) {
		t.Parallel()

		// Create taxonomy with nil internal map.
		tax := xbrl.NewTaxonomyForTest(nil)
		source := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{
			q2: c2,
		})

		tax.Merge(source)

		concepts := tax.Concepts()
		if assert.Len(t, concepts, 1) {
			got, ok := tax.Concept(q2)
			assert.True(t, ok)
			assert.Equal(t, "C2", got.ID())
		}
	})

	t.Run("merge with nil other", func(t *testing.T) {
		t.Parallel()

		tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{
			q1: c1,
		})

		tax.Merge(nil)

		concepts := tax.Concepts()
		if assert.Len(t, concepts, 1) {
			got, ok := tax.Concept(q1)
			assert.True(t, ok)
			assert.Equal(t, "C1", got.ID())
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()

		var nilTax *xbrl.Taxonomy
		other := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{
			q1: c1,
		})

		// Should not panic and should be a no-op.
		nilTax.Merge(other)
		assert.Nil(t, nilTax)
	})
}
