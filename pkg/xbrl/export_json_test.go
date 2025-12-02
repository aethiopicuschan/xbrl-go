package xbrl_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/aethiopicuschan/xbrl-go/pkg/xbrl"
	"github.com/stretchr/testify/assert"
)

// TestFactsAsJSONDTOs_NilDocument verifies that a nil *Document returns nil.
func TestFactsAsJSONDTOs_NilDocument(t *testing.T) {
	t.Parallel()

	var nilDoc *xbrl.Document

	dtos := nilDoc.FactsAsJSONDTOs()
	assert.Nil(t, dtos)
}

// TestFactsAsJSONDTOs_BasicBehavior checks conversion of facts to DTOs,
// including skipping nil facts and clearing Value when Nil=true.
func TestFactsAsJSONDTOs_BasicBehavior(t *testing.T) {
	t.Parallel()

	// QNames for names
	q1 := xbrl.NewQNameForTest("", "LocalOnly", "")
	q2 := xbrl.NewQNameForTest("p", "WithPrefix", "")
	q3 := xbrl.NewQNameForTest("p", "WithURI", "urn:ns")

	f1 := xbrl.NewFactForTest(
		xbrl.FactKindItem,
		q1,
		"v1",
		"C1",
		"U1",
		"",
		"",
		"F1",
		"",
		false,
	)
	f2 := xbrl.NewFactForTest(
		xbrl.FactKindItem,
		q2,
		"should be cleared when nil",
		"C2",
		"U2",
		"",
		"",
		"F2",
		"",
		true, // nil=true
	)
	f3 := xbrl.NewFactForTest(
		xbrl.FactKindItem,
		q3,
		"v3",
		"C3",
		"U3",
		"",
		"",
		"F3",
		"",
		false,
	)

	// Insert a nil fact in the slice; it should be skipped.
	doc := xbrl.NewDocumentForTest(nil, nil, nil, []*xbrl.Fact{f1, nil, f2, f3}, nil)

	dtos := doc.FactsAsJSONDTOs()

	if assert.Len(t, dtos, 3) {
		// f1
		assert.Equal(t, "LocalOnly", dtos[0].Name)
		assert.Equal(t, "v1", dtos[0].Value)
		assert.Equal(t, "C1", dtos[0].ContextRef)
		assert.Equal(t, "U1", dtos[0].UnitRef)
		assert.False(t, dtos[0].Nil)

		// f2 (nil fact -> value cleared)
		assert.Equal(t, "p:WithPrefix", dtos[1].Name)
		assert.Equal(t, "", dtos[1].Value)
		assert.Equal(t, "C2", dtos[1].ContextRef)
		assert.Equal(t, "U2", dtos[1].UnitRef)
		assert.True(t, dtos[1].Nil)

		// f3 (QName with URI -> curly-brace form)
		assert.Equal(t, "{urn:ns}WithURI", dtos[2].Name)
		assert.Equal(t, "v3", dtos[2].Value)
		assert.Equal(t, "C3", dtos[2].ContextRef)
		assert.Equal(t, "U3", dtos[2].UnitRef)
		assert.False(t, dtos[2].Nil)
	}
}

// TestEncodeFactsJSON_NilDocumentIsNoop verifies that EncodeFactsJSON on a nil
// *Document returns nil error and writes nothing.
func TestEncodeFactsJSON_NilDocumentIsNoop(t *testing.T) {
	t.Parallel()

	var nilDoc *xbrl.Document

	var buf bytes.Buffer
	err := nilDoc.EncodeFactsJSON(&buf, false)

	assert.NoError(t, err)
	assert.Equal(t, "", buf.String())
}

// TestEncodeFactsJSON_CompactAndPretty verifies JSON encoding behavior,
// including pretty-printing and disabled HTML escaping.
func TestEncodeFactsJSON_CompactAndPretty(t *testing.T) {
	t.Parallel()

	q := xbrl.NewQNameForTest("", "FactName", "")

	// Raw value with characters that are usually HTML-escaped.
	rawValue := `<tag>& "quote"`

	f1 := xbrl.NewFactForTest(
		xbrl.FactKindItem,
		q,
		rawValue,
		"C1",
		"U1",
		"",
		"",
		"F1",
		"en",
		false,
	)
	f2 := xbrl.NewFactForTest(
		xbrl.FactKindItem,
		q,
		"ignored when nil",
		"C2",
		"U2",
		"",
		"",
		"F2",
		"en",
		true, // nil=true
	)

	doc := xbrl.NewDocumentForTest(nil, nil, nil, []*xbrl.Fact{f1, f2}, nil)

	t.Run("compact JSON (pretty=false)", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		err := doc.EncodeFactsJSON(&buf, false)
		assert.NoError(t, err)

		// Decode to ensure valid JSON and correct structure.
		var got []xbrl.FactJSON
		err = json.Unmarshal(buf.Bytes(), &got)
		if assert.NoError(t, err) && assert.Len(t, got, 2) {
			// First fact
			assert.Equal(t, "FactName", got[0].Name)
			assert.Equal(t, rawValue, got[0].Value)
			assert.Equal(t, "C1", got[0].ContextRef)
			assert.Equal(t, "U1", got[0].UnitRef)
			assert.False(t, got[0].Nil)

			// Second fact (nil -> value must be empty)
			assert.Equal(t, "FactName", got[1].Name)
			assert.Equal(t, "", got[1].Value)
			assert.Equal(t, "C2", got[1].ContextRef)
			assert.Equal(t, "U2", got[1].UnitRef)
			assert.True(t, got[1].Nil)
		}

		// Ensure HTML characters are not escaped in the output JSON.
		s := buf.String()

		// "<" "&" should stay as-is
		assert.Contains(t, s, `<tag>&`)

		// Quotes are escaped in JSON as \"quote\"
		assert.Contains(t, s, `\"quote\"`)

		// And no \u003c / \u003e / \u0026 sequences
		assert.NotContains(t, s, `\u003c`)
		assert.NotContains(t, s, `\u003e`)
		assert.NotContains(t, s, `\u0026`)
	})

	t.Run("pretty JSON (pretty=true)", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		err := doc.EncodeFactsJSON(&buf, true)
		assert.NoError(t, err)

		s := buf.String()
		// Pretty JSON should contain newlines and indentation.
		assert.Contains(t, s, "\n  {")

		var got []xbrl.FactJSON
		err = json.Unmarshal([]byte(s), &got)
		if assert.NoError(t, err) && assert.Len(t, got, 2) {
			assert.Equal(t, "FactName", got[0].Name)
			assert.Equal(t, rawValue, got[0].Value)
			assert.Equal(t, "C1", got[0].ContextRef)
			assert.Equal(t, "U1", got[0].UnitRef)
			assert.False(t, got[0].Nil)

			assert.Equal(t, "FactName", got[1].Name)
			assert.Equal(t, "", got[1].Value)
			assert.Equal(t, "C2", got[1].ContextRef)
			assert.Equal(t, "U2", got[1].UnitRef)
			assert.True(t, got[1].Nil)
		}
	})
}
