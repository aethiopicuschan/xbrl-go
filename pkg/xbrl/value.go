package xbrl

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Namespaces commonly used in XBRL types.
const (
	nsXBRLI = "http://www.xbrl.org/2003/instance"
	nsXSD   = "http://www.w3.org/2001/XMLSchema"
)

// ConceptValueKind classifies the conceptual value type of a concept.
// This is a coarse-grained classification based on its @type.
type ConceptValueKind int

const (
	ConceptValueUnknown ConceptValueKind = iota
	ConceptValueString
	ConceptValueNumeric
	ConceptValueMonetary
	ConceptValueBoolean
	ConceptValueDate
	ConceptValueDateTime
)

// String implements fmt.Stringer.
func (k ConceptValueKind) String() string {
	switch k {
	case ConceptValueString:
		return "string"
	case ConceptValueNumeric:
		return "numeric"
	case ConceptValueMonetary:
		return "monetary"
	case ConceptValueBoolean:
		return "boolean"
	case ConceptValueDate:
		return "date"
	case ConceptValueDateTime:
		return "dateTime"
	default:
		return "unknown"
	}
}

// ValueKind returns a coarse-grained classification of the concept's
// value type, based on its @type QName.
//
// This function does not look at linkbases or custom types; it only
// inspects well-known XBRL and XML Schema types and falls back to
// ConceptValueString for unknown types.
func (c *Concept) ValueKind() ConceptValueKind {
	if c == nil {
		return ConceptValueUnknown
	}

	t := c.Type()
	uri := t.URI()
	local := t.Local()

	switch uri {
	case nsXBRLI:
		switch local {
		case "monetaryItemType":
			return ConceptValueMonetary
		case "sharesItemType", "perShareItemType",
			"decimalItemType", "integerItemType",
			"nonNegativeIntegerItemType", "nonPositiveIntegerItemType",
			"positiveIntegerItemType", "negativeIntegerItemType",
			"pureItemType", "fractionItemType":
			return ConceptValueNumeric
		case "booleanItemType":
			return ConceptValueBoolean
		case "dateItemType":
			return ConceptValueDate
		case "dateTimeItemType":
			return ConceptValueDateTime
		case "stringItemType":
			return ConceptValueString
		default:
			// Unknown xbrli type → treat as string.
			return ConceptValueString
		}
	case nsXSD:
		switch local {
		case "decimal", "integer", "nonNegativeInteger", "nonPositiveInteger",
			"positiveInteger", "negativeInteger", "int", "long", "short", "byte",
			"unsignedInt", "unsignedLong", "unsignedShort", "unsignedByte", "float", "double":
			return ConceptValueNumeric
		case "boolean":
			return ConceptValueBoolean
		case "date":
			return ConceptValueDate
		case "dateTime":
			return ConceptValueDateTime
		case "string", "normalizedString":
			return ConceptValueString
		default:
			return ConceptValueString
		}
	default:
		// Unknown namespace: be conservative and treat as string.
		return ConceptValueString
	}
}

// Errors returned by typed value helpers.
var (
	ErrNoTaxonomy      = errors.New("xbrl: no taxonomy attached to document")
	ErrNoConcept       = errors.New("xbrl: concept not found for fact")
	ErrUnsupportedType = errors.New("xbrl: unsupported value type for this conversion")
	ErrInvalidValue    = errors.New("xbrl: invalid lexical form for type")
)

// AsInt64 parses the fact's value as an int64, based on its concept type.
//
// The taxonomy must be attached to the Document (via SetTaxonomy or
// LoadTaxonomyFromSchemaRefs). The concept's ValueKind must be
// ConceptValueNumeric or ConceptValueMonetary.
func (d *Document) AsInt64(f *Fact) (int64, error) {
	if d == nil {
		return 0, fmt.Errorf("xbrl: document is nil")
	}
	if d.taxonomy == nil {
		return 0, ErrNoTaxonomy
	}
	if f == nil {
		return 0, fmt.Errorf("xbrl: fact is nil")
	}
	if f.IsNil() {
		return 0, ErrInvalidValue
	}

	c, ok := d.ConceptOf(f)
	if !ok || c == nil {
		return 0, ErrNoConcept
	}

	switch c.ValueKind() {
	case ConceptValueNumeric, ConceptValueMonetary:
		v := strings.TrimSpace(f.Value())
		if strings.ContainsAny(v, ".eE") {
			return 0, ErrInvalidValue
		}
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: %v", ErrInvalidValue, err)
		}
		return n, nil
	default:
		return 0, ErrUnsupportedType
	}
}

// AsFloat64 parses the fact's value as a float64, based on its concept type.
//
// The taxonomy must be attached to the Document. The concept's ValueKind
// must be ConceptValueNumeric or ConceptValueMonetary.
func (d *Document) AsFloat64(f *Fact) (float64, error) {
	if d == nil {
		return 0, fmt.Errorf("xbrl: document is nil")
	}
	if d.taxonomy == nil {
		return 0, ErrNoTaxonomy
	}
	if f == nil {
		return 0, fmt.Errorf("xbrl: fact is nil")
	}
	if f.IsNil() {
		return 0, ErrInvalidValue
	}

	c, ok := d.ConceptOf(f)
	if !ok || c == nil {
		return 0, ErrNoConcept
	}

	switch c.ValueKind() {
	case ConceptValueNumeric, ConceptValueMonetary:
		v := strings.TrimSpace(f.Value())
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: %v", ErrInvalidValue, err)
		}
		return n, nil
	default:
		return 0, ErrUnsupportedType
	}
}

// AsBool parses the fact's value as a bool, based on its concept type.
//
// The taxonomy must be attached and the concept's ValueKind must be
// ConceptValueBoolean.
//
// "true"/"1" → true, "false"/"0" → false
func (d *Document) AsBool(f *Fact) (bool, error) {
	if d == nil {
		return false, fmt.Errorf("xbrl: document is nil")
	}
	if d.taxonomy == nil {
		return false, ErrNoTaxonomy
	}
	if f == nil {
		return false, fmt.Errorf("xbrl: fact is nil")
	}
	if f.IsNil() {
		return false, ErrInvalidValue
	}

	c, ok := d.ConceptOf(f)
	if !ok || c == nil {
		return false, ErrNoConcept
	}

	if c.ValueKind() != ConceptValueBoolean {
		return false, ErrUnsupportedType
	}

	v := strings.TrimSpace(f.Value())
	switch strings.ToLower(v) {
	case "true", "1":
		return true, nil
	case "false", "0":
		return false, nil
	default:
		return false, ErrInvalidValue
	}
}

// AsTime parses the fact's value as time.Time, based on its concept type.
//
// The taxonomy must be attached and the concept's ValueKind must be
// ConceptValueDate or ConceptValueDateTime.
func (d *Document) AsTime(f *Fact, loc *time.Location) (time.Time, error) {
	if d == nil {
		return time.Time{}, fmt.Errorf("xbrl: document is nil")
	}
	if d.taxonomy == nil {
		return time.Time{}, ErrNoTaxonomy
	}
	if f == nil {
		return time.Time{}, fmt.Errorf("xbrl: fact is nil")
	}
	if f.IsNil() {
		return time.Time{}, ErrInvalidValue
	}

	c, ok := d.ConceptOf(f)
	if !ok || c == nil {
		return time.Time{}, ErrNoConcept
	}

	if loc == nil {
		loc = time.UTC
	}

	v := strings.TrimSpace(f.Value())

	switch c.ValueKind() {
	case ConceptValueDate:
		// ISO 8601 yyyy-mm-dd
		t, err := time.ParseInLocation("2006-01-02", v, loc)
		if err != nil {
			return time.Time{}, fmt.Errorf("%w: %v", ErrInvalidValue, err)
		}
		return t, nil
	case ConceptValueDateTime:
		// Try RFC3339
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t.In(loc), nil
		}
		// Allow yyyy-mm-ddThh:mm:ss without timezone
		if t, err := time.ParseInLocation("2006-01-02T15:04:05", v, loc); err == nil {
			return t, nil
		}
		return time.Time{}, ErrInvalidValue
	default:
		return time.Time{}, ErrUnsupportedType
	}
}
