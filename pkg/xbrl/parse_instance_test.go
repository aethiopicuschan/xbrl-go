package xbrl_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aethiopicuschan/xbrl-go/pkg/xbrl"
)

const minimalInstance = `
<xbrli:xbrl
    xmlns:xbrli="http://www.xbrl.org/2003/instance"
    xmlns:link="http://www.xbrl.org/2003/linkbase"
    xmlns:ex="http://example.com/xbrl">
  <xbrli:schemaRef xlink:type="simple"
    xlink:href="http://example.com/schema.xsd"
    xmlns:xlink="http://www.w3.org/1999/xlink"/>
  <xbrli:context id="C1">
    <xbrli:entity>
      <xbrli:identifier scheme="http://example.com/entity">ABC</xbrli:identifier>
    </xbrli:entity>
    <xbrli:period>
      <xbrli:instant>2025-01-01</xbrli:instant>
    </xbrli:period>
  </xbrli:context>
  <xbrli:unit id="U1">
    <xbrli:measure>iso4217:JPY</xbrli:measure>
  </xbrli:unit>
  <ex:Revenue contextRef="C1" unitRef="U1" decimals="0">12345</ex:Revenue>
</xbrli:xbrl>
`

const extendedInstance = `<?xml version="1.0" encoding="utf-8"?>
<xbrli:xbrl
    xmlns:xbrli="http://www.xbrl.org/2003/instance"
    xmlns:link="http://www.xbrl.org/2003/linkbase"
    xmlns:ex="http://example.com/xbrl"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xmlns:iso4217="urn:iso:std:iso:4217">

  <xbrli:schemaRef xlink:type="simple"
    xlink:href="http://example.com/schema.xsd"
    xmlns:xlink="http://www.w3.org/1999/xlink"/>

  <!-- C1: entity + segment(explicitMember) + scenario(typedMember) + start/end period -->
  <xbrli:context id="C1">
    <xbrli:entity>
      <xbrli:identifier scheme="http://example.com/entity">
        ABC
      </xbrli:identifier>
      <xbrli:segment>
        <xbrli:explicitMember dimension="ex:Region">
          ex:Japan
        </xbrli:explicitMember>
      </xbrli:segment>
    </xbrli:entity>
    <xbrli:period>
      <xbrli:startDate>2025-01-01</xbrli:startDate>
      <xbrli:endDate>2025-12-31</xbrli:endDate>
    </xbrli:period>
    <xbrli:scenario>
      <xbrli:typedMember dimension="ex:Scenario">
        <ex:ScenarioType> Base </ex:ScenarioType>
      </xbrli:typedMember>
    </xbrli:scenario>
  </xbrli:context>

  <!-- C2: forever period -->
  <xbrli:context id="C2">
    <xbrli:entity>
      <xbrli:identifier scheme="http://example.com/entity">
        XYZ
      </xbrli:identifier>
    </xbrli:entity>
    <xbrli:period>
      <xbrli:forever/>
    </xbrli:period>
  </xbrli:context>

  <!-- unit: simple -->
  <xbrli:unit id="U1">
    <xbrli:measure>iso4217:JPY</xbrli:measure>
  </xbrli:unit>

  <!-- unit: divide -->
  <xbrli:unit id="Udiv">
    <xbrli:divide>
      <xbrli:unitNumerator>
        <xbrli:measure>iso4217:JPY</xbrli:measure>
      </xbrli:unitNumerator>
      <xbrli:unitDenominator>
        <xbrli:measure>iso4217:USD</xbrli:measure>
      </xbrli:unitDenominator>
    </xbrli:divide>
  </xbrli:unit>

  <xbrli:unit id="U2" xmlns="urn:iso:std:iso:4217">
    <xbrli:measure>JPY</xbrli:measure>
  </xbrli:unit>

  <ex:Revenue contextRef="C1" unitRef="U1" decimals="0" precision="2" id="F1" lang="ja">
    12345
  </ex:Revenue>

  <ex:NilFact contextRef="C1" xsi:nil="true"/>

  <ex:OtherElement>no contextRef, should be ignored</ex:OtherElement>
</xbrli:xbrl>
`

func TestParse_MinimalInstance_Counts(t *testing.T) {
	t.Parallel()

	doc, err := xbrl.Parse(strings.NewReader(minimalInstance))
	assert.NoError(t, err)

	assert.Len(t, doc.SchemaRefs(), 1)
	assert.Len(t, doc.Contexts(), 1)
	assert.Len(t, doc.Units(), 1)
	assert.Len(t, doc.Facts(), 1)
}

func TestParse_MinimalInstance_FactBasics(t *testing.T) {
	t.Parallel()

	doc, err := xbrl.Parse(strings.NewReader(minimalInstance))
	assert.NoError(t, err)

	facts := doc.Facts()
	if assert.Len(t, facts, 1) {
		f := facts[0]
		assert.NotNil(t, f)
	}
}

func TestParse_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		xml         string
		expectError string // substring that should appear in error message
	}{
		{
			name: "Malformed XML root tag",
			xml: `
				<xbrli:xbrl xmlns:xbrli="http://www.xbrl.org/2003/instance"
			`, // missing closing '>' and closing tags
			expectError: "xbrl: decode token", // from Parse when xml.Decoder.Token fails
		},
		{
			name: "Unexpected EOF inside context",
			xml: `
				<xbrli:xbrl xmlns:xbrli="http://www.xbrl.org/2003/instance">
				  <xbrli:context id="C1">
					<xbrli:entity>
					  <xbrli:identifier scheme="http://example.com/entity">ABC</xbrli:identifier>
					</xbrli:entity>
					<xbrli:period>
					  <xbrli:instant>2025-01-01</xbrli:instant>
					</xbrli:period>
				  <!-- missing </xbrli:context> and </xbrli:xbrl> -->
				`,
			expectError: "parse context", // from parseContext
		},
		{
			name: "Unexpected EOF inside unit",
			xml: `
				<xbrli:xbrl xmlns:xbrli="http://www.xbrl.org/2003/instance">
				  <xbrli:unit id="U1">
					<xbrli:measure>iso4217:JPY</xbrli:measure>
				  <!-- missing </xbrli:unit> and </xbrli:xbrl> -->
				`,
			expectError: "parse unit", // from parseUnit or generic decode error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := xbrl.Parse(strings.NewReader(tt.xml))
			if assert.Error(t, err, "Parse should fail for malformed input") {
				assert.Contains(t, err.Error(), tt.expectError)
			}
		})
	}
}

func TestParseFile_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := xbrl.ParseFile("this-file-does-not-exist-12345.xbrl")
	if assert.Error(t, err, "ParseFile should fail for non-existent file") {
		assert.Contains(t, err.Error(), "xbrl: open file")
	}
}

func TestParse_ExtendedInstance_StructureAndEncoding(t *testing.T) {
	t.Parallel()

	doc, err := xbrl.Parse(strings.NewReader(extendedInstance))
	require.NoError(t, err)

	// 2 schemaRefs
	schemaRefs := doc.SchemaRefs()
	assert.Len(t, schemaRefs, 1)

	// 2 contexts（C1, C2）
	ctxs := doc.Contexts()
	if assert.Len(t, ctxs, 2) {
		_, ok1 := ctxs["C1"]
		_, ok2 := ctxs["C2"]
		assert.True(t, ok1, "context C1 should exist")
		assert.True(t, ok2, "context C2 should exist")
	}

	// 3 units（U1, Udiv, U2）
	units := doc.Units()
	if assert.Len(t, units, 3) {
		_, ok1 := units["U1"]
		_, okDiv := units["Udiv"]
		_, ok2 := units["U2"]
		assert.True(t, ok1, "unit U1 should exist")
		assert.True(t, okDiv, "unit Udiv should exist")
		assert.True(t, ok2, "unit U2 should exist")
	}

	// 2 facts（Revenue, NilFact）
	facts := doc.Facts()
	assert.Len(t, facts, 2)
}

func TestParse_ElementWithoutContextRef_IsIgnoredAsFact(t *testing.T) {
	t.Parallel()

	xmlStr := `
	<xbrli:xbrl
	    xmlns:xbrli="http://www.xbrl.org/2003/instance"
	    xmlns:ex="http://example.com/xbrl">
	  <xbrli:context id="C1">
	    <xbrli:entity>
	      <xbrli:identifier scheme="http://example.com/entity">ABC</xbrli:identifier>
	    </xbrli:entity>
	    <xbrli:period>
	      <xbrli:instant>2025-01-01</xbrli:instant>
	    </xbrli:period>
	  </xbrli:context>
	  <ex:Something>no contextRef</ex:Something>
	</xbrli:xbrl>
	`

	doc, err := xbrl.Parse(strings.NewReader(xmlStr))
	require.NoError(t, err)

	assert.Len(t, doc.Contexts(), 1)
	assert.Len(t, doc.Facts(), 0, "Element without contextRef should not be parsed as Fact")
}

func TestParseFile_Success(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "instance.xbrl")

	err := os.WriteFile(path, []byte(minimalInstance), 0o644)
	require.NoError(t, err)

	doc, err := xbrl.ParseFile(path)
	require.NoError(t, err)
	require.NotNil(t, doc)

	assert.Len(t, doc.SchemaRefs(), 1)
	assert.Len(t, doc.Contexts(), 1)
	assert.Len(t, doc.Units(), 1)
	assert.Len(t, doc.Facts(), 1)
}
