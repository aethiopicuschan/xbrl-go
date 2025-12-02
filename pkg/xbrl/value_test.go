package xbrl_test

import (
	"errors"
	"testing"
	"time"

	"github.com/aethiopicuschan/xbrl-go/pkg/xbrl"
	"github.com/stretchr/testify/assert"
)

const (
	nsXBRLI = "http://www.xbrl.org/2003/instance"
	nsXSD   = "http://www.w3.org/2001/XMLSchema"
)

func newDocFactWithType(t testing.TB, typeURI, typeLocal, value string, kind xbrl.ConceptValueKind) (*xbrl.Document, *xbrl.Fact) {
	t.Helper()

	q := xbrl.NewQNameForTest("x", "TestConcept", "http://example.com")

	typeQName := xbrl.NewQNameForTest("xbrli", typeLocal, typeURI)

	concept := xbrl.NewConceptForTest(
		q,
		"TestConceptID",
		xbrl.NewQNameForTest("", "", ""),
		typeQName,
		false,
		false,
		"",
		"",
	)

	// Taxonomy
	tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{
		q: concept,
	})

	// Fact
	fact := xbrl.NewFactForTest(
		0,       // 0
		q,       // name
		value,   // value
		"ctx1",  // contextRef
		"",      // unitRef
		"",      // decimals
		"",      // precision
		"fact1", // id
		"ja",    // lang
		false,   // isNil
	)

	// Document
	doc := xbrl.NewDocumentForTest(
		nil,
		map[string]*xbrl.Context{},
		map[string]*xbrl.Unit{},
		[]*xbrl.Fact{fact},
		tax,
	)

	assert.Equal(t, kind, concept.ValueKind())

	return doc, fact
}

//------------------------------------------------------------
// ConceptValueKind.String
//------------------------------------------------------------

func TestConceptValueKind_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind xbrl.ConceptValueKind
		want string
	}{
		{"Unknown", xbrl.ConceptValueUnknown, "unknown"},
		{"String", xbrl.ConceptValueString, "string"},
		{"Numeric", xbrl.ConceptValueNumeric, "numeric"},
		{"Monetary", xbrl.ConceptValueMonetary, "monetary"},
		{"Boolean", xbrl.ConceptValueBoolean, "boolean"},
		{"Date", xbrl.ConceptValueDate, "date"},
		{"DateTime", xbrl.ConceptValueDateTime, "dateTime"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.kind.String())
		})
	}
}

//------------------------------------------------------------
// (*Concept).ValueKind
//------------------------------------------------------------

func TestConcept_ValueKind(t *testing.T) {
	t.Parallel()

	type args struct {
		typeURI   string
		typeLocal string
	}
	tests := []struct {
		name string
		args args
		want xbrl.ConceptValueKind
	}{
		// nsXBRLI
		{"XBRLI_Monetary", args{nsXBRLI, "monetaryItemType"}, xbrl.ConceptValueMonetary},
		{"XBRLI_NumericInteger", args{nsXBRLI, "integerItemType"}, xbrl.ConceptValueNumeric},
		{"XBRLI_Shares", args{nsXBRLI, "sharesItemType"}, xbrl.ConceptValueNumeric},
		{"XBRLI_Boolean", args{nsXBRLI, "booleanItemType"}, xbrl.ConceptValueBoolean},
		{"XBRLI_Date", args{nsXBRLI, "dateItemType"}, xbrl.ConceptValueDate},
		{"XBRLI_DateTime", args{nsXBRLI, "dateTimeItemType"}, xbrl.ConceptValueDateTime},
		{"XBRLI_String", args{nsXBRLI, "stringItemType"}, xbrl.ConceptValueString},
		{"XBRLI_UnknownLocal", args{nsXBRLI, "unknownItemType"}, xbrl.ConceptValueString},

		// nsXSD
		{"XSD_Decimal", args{nsXSD, "decimal"}, xbrl.ConceptValueNumeric},
		{"XSD_Integer", args{nsXSD, "integer"}, xbrl.ConceptValueNumeric},
		{"XSD_Boolean", args{nsXSD, "boolean"}, xbrl.ConceptValueBoolean},
		{"XSD_Date", args{nsXSD, "date"}, xbrl.ConceptValueDate},
		{"XSD_DateTime", args{nsXSD, "dateTime"}, xbrl.ConceptValueDateTime},
		{"XSD_String", args{nsXSD, "string"}, xbrl.ConceptValueString},
		{"XSD_NormalizedString", args{nsXSD, "normalizedString"}, xbrl.ConceptValueString},
		{"XSD_UnknownLocal", args{nsXSD, "someType"}, xbrl.ConceptValueString},

		// Unknown namespace
		{"UnknownNamespace", args{"http://example.com", "any"}, xbrl.ConceptValueString},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			q := xbrl.NewQNameForTest("x", "Concept", "http://example.com")
			typeQName := xbrl.NewQNameForTest("t", tc.args.typeLocal, tc.args.typeURI)

			concept := xbrl.NewConceptForTest(
				q,
				"id",
				xbrl.NewQNameForTest("", "", ""),
				typeQName,
				false,
				false,
				"",
				"",
			)

			assert.Equal(t, tc.want, concept.ValueKind())
		})
	}

	t.Run("NilConcept", func(t *testing.T) {
		t.Parallel()
		var c *xbrl.Concept
		assert.Equal(t, xbrl.ConceptValueUnknown, c.ValueKind())
	})
}

//------------------------------------------------------------
// Document.AsInt64
//------------------------------------------------------------

func TestDocument_AsInt64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) (*xbrl.Document, *xbrl.Fact)
		want    int64
		wantErr error
		checkIs func(error) bool
	}{
		{
			name: "NilDocument",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				return nil, nil
			},
			want:    0,
			wantErr: errors.New("xbrl: document is nil"),
		},
		{
			name: "NoTaxonomy",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc := xbrl.NewDocumentForTest(nil, nil, nil, nil, nil)
				f := xbrl.NewFactForTest(0, xbrl.NewQNameForTest("", "n", ""), "123", "ctx", "", "", "", "id", "", false)
				return doc, f
			},
			wantErr: xbrl.ErrNoTaxonomy,
		},
		{
			name: "NilFact",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{})
				doc := xbrl.NewDocumentForTest(nil, nil, nil, nil, tax)
				return doc, nil
			},
			wantErr: errors.New("xbrl: fact is nil"),
		},
		{
			name: "NilFactValue",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				q := xbrl.NewQNameForTest("x", "c", "http://example.com")
				typeQName := xbrl.NewQNameForTest("t", "integerItemType", nsXBRLI)
				concept := xbrl.NewConceptForTest(q, "id", xbrl.NewQNameForTest("", "", ""), typeQName, false, false, "", "")
				tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{q: concept})

				f := xbrl.NewFactForTest(0, q, "123", "ctx", "", "", "", "id", "", true)
				doc := xbrl.NewDocumentForTest(nil, nil, nil, []*xbrl.Fact{f}, tax)
				return doc, f
			},
			wantErr: xbrl.ErrInvalidValue,
		},
		{
			name: "NoConcept",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				q := xbrl.NewQNameForTest("x", "c", "http://example.com")
				tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{})
				f := xbrl.NewFactForTest(0, q, "123", "ctx", "", "", "", "id", "", false)
				doc := xbrl.NewDocumentForTest(nil, nil, nil, []*xbrl.Fact{f}, tax)
				return doc, f
			},
			wantErr: xbrl.ErrNoConcept,
		},
		{
			name: "UnsupportedType",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				q := xbrl.NewQNameForTest("x", "c", "http://example.com")
				typeQName := xbrl.NewQNameForTest("t", "boolean", nsXSD)
				concept := xbrl.NewConceptForTest(q, "id", xbrl.NewQNameForTest("", "", ""), typeQName, false, false, "", "")
				tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{q: concept})
				f := xbrl.NewFactForTest(0, q, "1", "ctx", "", "", "", "id", "", false)
				doc := xbrl.NewDocumentForTest(nil, nil, nil, []*xbrl.Fact{f}, tax)
				return doc, f
			},
			wantErr: xbrl.ErrUnsupportedType,
		},
		{
			name: "InvalidDecimalForm",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc, f := newDocFactWithType(t, nsXSD, "integer", "123.45", xbrl.ConceptValueNumeric)
				return doc, f
			},
			wantErr: xbrl.ErrInvalidValue,
		},
		{
			name: "ParseError",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc, f := newDocFactWithType(t, nsXSD, "integer", "not-an-int", xbrl.ConceptValueNumeric)
				return doc, f
			},
			checkIs: func(err error) bool {
				return errors.Is(err, xbrl.ErrInvalidValue)
			},
		},
		{
			name: "OK_Numeric",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc, f := newDocFactWithType(t, nsXSD, "integer", "  42 ", xbrl.ConceptValueNumeric)
				return doc, f
			},
			want: 42,
		},
		{
			name: "OK_Monetary",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc, f := newDocFactWithType(t, nsXBRLI, "monetaryItemType", "1000", xbrl.ConceptValueMonetary)
				return doc, f
			},
			want: 1000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, fact := tc.setup(t)

			var got int64
			var err error

			if doc == nil {
				var d *xbrl.Document
				got, err = d.AsInt64(fact)
			} else {
				got, err = doc.AsInt64(fact)
			}

			if tc.checkIs != nil {
				assert.True(t, tc.checkIs(err), "error = %v", err)
				return
			}

			if tc.wantErr != nil {
				if msg := tc.wantErr.Error(); msg != "" {
					assert.EqualError(t, err, msg)
				} else {
					assert.ErrorIs(t, err, tc.wantErr)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

//------------------------------------------------------------
// Document.AsFloat64
//------------------------------------------------------------

func TestDocument_AsFloat64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) (*xbrl.Document, *xbrl.Fact)
		want    float64
		wantErr error
		checkIs func(error) bool
	}{
		{
			name: "NilDocument",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				return nil, nil
			},
			wantErr: errors.New("xbrl: document is nil"),
		},
		{
			name: "NoTaxonomy",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc := xbrl.NewDocumentForTest(nil, nil, nil, nil, nil)
				f := xbrl.NewFactForTest(0, xbrl.NewQNameForTest("", "n", ""), "123.4", "ctx", "", "", "", "id", "", false)
				return doc, f
			},
			wantErr: xbrl.ErrNoTaxonomy,
		},
		{
			name: "NilFact",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{})
				doc := xbrl.NewDocumentForTest(nil, nil, nil, nil, tax)
				return doc, nil
			},
			wantErr: errors.New("xbrl: fact is nil"),
		},
		{
			name: "NilFactValue",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				q := xbrl.NewQNameForTest("x", "c", "http://example.com")
				typeQName := xbrl.NewQNameForTest("t", "decimal", nsXSD)
				concept := xbrl.NewConceptForTest(q, "id", xbrl.NewQNameForTest("", "", ""), typeQName, false, false, "", "")
				tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{q: concept})
				f := xbrl.NewFactForTest(0, q, "1.23", "ctx", "", "", "", "id", "", true)
				doc := xbrl.NewDocumentForTest(nil, nil, nil, []*xbrl.Fact{f}, tax)
				return doc, f
			},
			wantErr: xbrl.ErrInvalidValue,
		},
		{
			name: "NoConcept",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				q := xbrl.NewQNameForTest("x", "c", "http://example.com")
				tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{})
				f := xbrl.NewFactForTest(0, q, "1.23", "ctx", "", "", "", "id", "", false)
				doc := xbrl.NewDocumentForTest(nil, nil, nil, []*xbrl.Fact{f}, tax)
				return doc, f
			},
			wantErr: xbrl.ErrNoConcept,
		},
		{
			name: "UnsupportedType",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				q := xbrl.NewQNameForTest("x", "c", "http://example.com")
				typeQName := xbrl.NewQNameForTest("t", "boolean", nsXSD)
				concept := xbrl.NewConceptForTest(q, "id", xbrl.NewQNameForTest("", "", ""), typeQName, false, false, "", "")
				tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{q: concept})
				f := xbrl.NewFactForTest(0, q, "true", "ctx", "", "", "", "id", "", false)
				doc := xbrl.NewDocumentForTest(nil, nil, nil, []*xbrl.Fact{f}, tax)
				return doc, f
			},
			wantErr: xbrl.ErrUnsupportedType,
		},
		{
			name: "ParseError",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc, f := newDocFactWithType(t, nsXSD, "decimal", "not-a-float", xbrl.ConceptValueNumeric)
				return doc, f
			},
			checkIs: func(err error) bool {
				return errors.Is(err, xbrl.ErrInvalidValue)
			},
		},
		{
			name: "OK_Numeric",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc, f := newDocFactWithType(t, nsXSD, "decimal", "  123.45 ", xbrl.ConceptValueNumeric)
				return doc, f
			},
			want: 123.45,
		},
		{
			name: "OK_Monetary",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc, f := newDocFactWithType(t, nsXBRLI, "monetaryItemType", "1000.5", xbrl.ConceptValueMonetary)
				return doc, f
			},
			want: 1000.5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, fact := tc.setup(t)

			var got float64
			var err error

			if doc == nil {
				var d *xbrl.Document
				got, err = d.AsFloat64(fact)
			} else {
				got, err = doc.AsFloat64(fact)
			}

			if tc.checkIs != nil {
				assert.True(t, tc.checkIs(err), "error = %v", err)
				return
			}

			if tc.wantErr != nil {
				if msg := tc.wantErr.Error(); msg != "" {
					assert.EqualError(t, err, msg)
				} else {
					assert.ErrorIs(t, err, tc.wantErr)
				}
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tc.want, got, 1e-9)
			}
		})
	}
}

//------------------------------------------------------------
// Document.AsBool
//------------------------------------------------------------

func TestDocument_AsBool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) (*xbrl.Document, *xbrl.Fact)
		want    bool
		wantErr error
	}{
		{
			name: "NilDocument",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				return nil, nil
			},
			wantErr: errors.New("xbrl: document is nil"),
		},
		{
			name: "NoTaxonomy",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc := xbrl.NewDocumentForTest(nil, nil, nil, nil, nil)
				f := xbrl.NewFactForTest(0, xbrl.NewQNameForTest("", "n", ""), "true", "ctx", "", "", "", "id", "", false)
				return doc, f
			},
			wantErr: xbrl.ErrNoTaxonomy,
		},
		{
			name: "NilFact",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{})
				doc := xbrl.NewDocumentForTest(nil, nil, nil, nil, tax)
				return doc, nil
			},
			wantErr: errors.New("xbrl: fact is nil"),
		},
		{
			name: "NilFactValue",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				q := xbrl.NewQNameForTest("x", "c", "http://example.com")
				typeQName := xbrl.NewQNameForTest("t", "boolean", nsXSD)
				concept := xbrl.NewConceptForTest(q, "id", xbrl.NewQNameForTest("", "", ""), typeQName, false, false, "", "")
				tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{q: concept})
				f := xbrl.NewFactForTest(0, q, "true", "ctx", "", "", "", "id", "", true)
				doc := xbrl.NewDocumentForTest(nil, nil, nil, []*xbrl.Fact{f}, tax)
				return doc, f
			},
			wantErr: xbrl.ErrInvalidValue,
		},
		{
			name: "NoConcept",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				q := xbrl.NewQNameForTest("x", "c", "http://example.com")
				tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{})
				f := xbrl.NewFactForTest(0, q, "true", "ctx", "", "", "", "id", "", false)
				doc := xbrl.NewDocumentForTest(nil, nil, nil, []*xbrl.Fact{f}, tax)
				return doc, f
			},
			wantErr: xbrl.ErrNoConcept,
		},
		{
			name: "UnsupportedType",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				// numeric → unsupported
				doc, f := newDocFactWithType(t, nsXSD, "integer", "1", xbrl.ConceptValueNumeric)
				return doc, f
			},
			wantErr: xbrl.ErrUnsupportedType,
		},
		{
			name: "InvalidLexical",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc, f := newDocFactWithType(t, nsXSD, "boolean", "yes", xbrl.ConceptValueBoolean)
				return doc, f
			},
			wantErr: xbrl.ErrInvalidValue,
		},
		{
			name: "TrueVariants",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc, f := newDocFactWithType(t, nsXSD, "boolean", "  True ", xbrl.ConceptValueBoolean)
				return doc, f
			},
			want: true,
		},
		{
			name: "FalseVariants",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc, f := newDocFactWithType(t, nsXSD, "boolean", "0", xbrl.ConceptValueBoolean)
				return doc, f
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, fact := tc.setup(t)

			var got bool
			var err error

			if doc == nil {
				var d *xbrl.Document
				got, err = d.AsBool(fact)
			} else {
				got, err = doc.AsBool(fact)
			}

			if tc.wantErr != nil {
				if msg := tc.wantErr.Error(); msg != "" {
					assert.EqualError(t, err, msg)
				} else {
					assert.ErrorIs(t, err, tc.wantErr)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

// ------------------------------------------------------------
// Document.AsTime
// ------------------------------------------------------------

func TestDocument_AsTime(t *testing.T) {
	t.Parallel()

	jst := time.FixedZone("JST", 9*60*60)

	tests := []struct {
		name       string
		setup      func(t *testing.T) (*xbrl.Document, *xbrl.Fact)
		loc        *time.Location
		want       time.Time
		wantErr    error
		wantErrMsg string
		checkIs    func(error) bool
	}{
		{
			name: "NilDocument",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				return nil, nil
			},
			loc:        time.UTC,
			wantErrMsg: "xbrl: document is nil",
		},
		{
			name: "NoTaxonomy",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc := xbrl.NewDocumentForTest(nil, nil, nil, nil, nil)
				f := xbrl.NewFactForTest(0, xbrl.NewQNameForTest("", "n", ""), "2025-01-02", "ctx", "", "", "", "id", "", false)
				return doc, f
			},
			loc:     time.UTC,
			wantErr: xbrl.ErrNoTaxonomy,
		},
		{
			name: "NilFact",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{})
				doc := xbrl.NewDocumentForTest(nil, nil, nil, nil, tax)
				return doc, nil
			},
			loc:        time.UTC,
			wantErrMsg: "xbrl: fact is nil",
		},
		{
			name: "NilFactValue",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				q := xbrl.NewQNameForTest("x", "c", "http://example.com")
				typeQName := xbrl.NewQNameForTest("t", "date", nsXSD)
				concept := xbrl.NewConceptForTest(q, "id", xbrl.NewQNameForTest("", "", ""), typeQName, false, false, "", "")
				tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{q: concept})
				f := xbrl.NewFactForTest(0, q, "2025-01-02", "ctx", "", "", "", "id", "", true)
				doc := xbrl.NewDocumentForTest(nil, nil, nil, []*xbrl.Fact{f}, tax)
				return doc, f
			},
			loc:     time.UTC,
			wantErr: xbrl.ErrInvalidValue,
		},
		{
			name: "NoConcept",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				q := xbrl.NewQNameForTest("x", "c", "http://example.com")
				tax := xbrl.NewTaxonomyForTest(map[xbrl.QName]*xbrl.Concept{})
				f := xbrl.NewFactForTest(0, q, "2025-01-02", "ctx", "", "", "", "id", "", false)
				doc := xbrl.NewDocumentForTest(nil, nil, nil, []*xbrl.Fact{f}, tax)
				return doc, f
			},
			loc:     time.UTC,
			wantErr: xbrl.ErrNoConcept,
		},
		{
			name: "UnsupportedType",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				// boolean → unsupported
				doc, f := newDocFactWithType(t, nsXSD, "boolean", "true", xbrl.ConceptValueBoolean)
				return doc, f
			},
			loc:     time.UTC,
			wantErr: xbrl.ErrUnsupportedType,
		},
		{
			name: "Date_OK_UTC_defaultLoc",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc, f := newDocFactWithType(t, nsXSD, "date", "2025-01-02", xbrl.ConceptValueDate)
				return doc, f
			},
			loc:  nil, // nil → UTC
			want: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Date_Invalid",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc, f := newDocFactWithType(t, nsXSD, "date", "2025/01/02", xbrl.ConceptValueDate)
				return doc, f
			},
			loc: jst,
			checkIs: func(err error) bool {
				return errors.Is(err, xbrl.ErrInvalidValue)
			},
		},
		{
			name: "DateTime_RFC3339_OK",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc, f := newDocFactWithType(t, nsXSD, "dateTime", "2025-01-02T15:04:05Z", xbrl.ConceptValueDateTime)
				return doc, f
			},
			loc: jst,
			want: func() time.Time {
				tm, _ := time.Parse(time.RFC3339, "2025-01-02T15:04:05Z")
				return tm.In(jst)
			}(),
		},
		{
			name: "DateTime_NoTZ_OK",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc, f := newDocFactWithType(t, nsXSD, "dateTime", "2025-01-02T15:04:05", xbrl.ConceptValueDateTime)
				return doc, f
			},
			loc:  jst,
			want: time.Date(2025, 1, 2, 15, 4, 5, 0, jst),
		},
		{
			name: "DateTime_Invalid",
			setup: func(t *testing.T) (*xbrl.Document, *xbrl.Fact) {
				doc, f := newDocFactWithType(t, nsXSD, "dateTime", "invalid", xbrl.ConceptValueDateTime)
				return doc, f
			},
			loc:     jst,
			wantErr: xbrl.ErrInvalidValue,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, fact := tc.setup(t)

			var got time.Time
			var err error

			if doc == nil {
				var d *xbrl.Document
				got, err = d.AsTime(fact, tc.loc)
			} else {
				got, err = doc.AsTime(fact, tc.loc)
			}

			if tc.checkIs != nil {
				assert.True(t, tc.checkIs(err), "error = %v", err)
				return
			}

			if tc.wantErrMsg != "" {
				assert.EqualError(t, err, tc.wantErrMsg)
				return
			}

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.True(t, got.Equal(tc.want), "got=%v want=%v", got, tc.want)
				assert.Equal(t, tc.want.Location(), got.Location())
			}
		})
	}
}
