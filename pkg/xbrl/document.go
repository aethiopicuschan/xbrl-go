package xbrl

import (
	"fmt"
	"io"
	"maps"
)

// Document represents a parsed XBRL instance document.
type Document struct {
	schemaRefs []SchemaRef
	contexts   map[string]*Context
	units      map[string]*Unit
	facts      []*Fact
	taxonomy   *Taxonomy
}

// SchemaRef represents a <schemaRef> element in an XBRL instance.
type SchemaRef struct {
	href string
}

// Context represents an XBRL <context> element.
type Context struct {
	id         string
	entity     Entity
	period     Period
	dimensions []Dimension
}

// Entity represents the <entity> of a context.
type Entity struct {
	identifier ContextIdentifier
}

// ContextIdentifier represents <identifier> inside <entity>.
type ContextIdentifier struct {
	scheme string
	value  string
}

// Period represents the <period> of a context.
type Period struct {
	instant   *string
	startDate *string
	endDate   *string
	forever   bool
}

// Unit represents an XBRL <unit> element.
//
// There are two major forms:
//   - simple unit: <unit><measure>...</measure></unit>
//   - divide unit: <unit><divide><unitNumerator>...</unitNumerator><unitDenominator>...</unitDenominator></divide></unit>
//
// For simple units, Measures() returns the measures and IsDivide() is false.
// For divide units, IsDivide() is true and NumeratorMeasures()/DenominatorMeasures()
// return the measures in numerator/denominator respectively.
type Unit struct {
	id string

	// simple measures (top-level <measure>)
	measures []QName

	// divide unit
	divide      bool
	numerator   []QName
	denominator []QName
}

// QName represents a qualified name with prefix, local name, and URI.
type QName struct {
	prefix string
	local  string
	uri    string
}

// FactKind describes the kind of fact.
type FactKind int

const (
	FactKindUnknown FactKind = iota
	FactKindItem
)

// Fact represents a single XBRL fact (item).
type Fact struct {
	kind FactKind

	name QName

	value string

	contextRef string
	unitRef    string
	decimals   string
	precision  string
	id         string
	lang       string
	nil        bool
}

// Dimension represents a dimensional qualifier (explicit or typed)
// attached to a context via <segment> or <scenario>.
type Dimension struct {
	dimension  QName  // the dimension QName from the "dimension" attribute
	explicit   bool   // true for explicitMember, false for typedMember
	member     QName  // explicit member QName (zero value if typed)
	typedValue string // raw inner XML for typedMember (empty for explicit)
}

// Dimension returns the QName of the dimension (the @dimension attribute).
func (d Dimension) Dimension() QName {
	return d.dimension
}

// IsExplicit reports whether this is an explicit dimension.
func (d Dimension) IsExplicit() bool {
	return d.explicit
}

// Member returns the explicit member QName.
//
// For typed dimensions this returns the zero value.
func (d Dimension) Member() QName {
	return d.member
}

// TypedValue returns the raw inner XML for a typed dimension.
//
// For explicit dimensions this returns an empty string.
func (d Dimension) TypedValue() string {
	return d.typedValue
}

// SchemaRefs returns a copy of the schema references in the document.
func (d *Document) SchemaRefs() []SchemaRef {
	if d == nil {
		return nil
	}
	out := make([]SchemaRef, len(d.schemaRefs))
	copy(out, d.schemaRefs)
	return out
}

// Contexts returns a copy of the contexts in the document.
func (d *Document) Contexts() map[string]*Context {
	if d == nil {
		return nil
	}
	out := make(map[string]*Context, len(d.contexts))
	maps.Copy(out, d.contexts)
	return out
}

// Units returns a copy of the units in the document.
func (d *Document) Units() map[string]*Unit {
	if d == nil {
		return nil
	}
	out := make(map[string]*Unit, len(d.units))
	maps.Copy(out, d.units)
	return out
}

// Facts returns a copy of the facts in the document.
func (d *Document) Facts() []*Fact {
	if d == nil {
		return nil
	}
	out := make([]*Fact, len(d.facts))
	copy(out, d.facts)
	return out
}

// ContextByID returns the context with the given ID, if present.
func (d *Document) ContextByID(id string) (*Context, bool) {
	if d == nil {
		return nil, false
	}
	ctx, ok := d.contexts[id]
	return ctx, ok
}

// UnitByID returns the unit with the given ID, if present.
func (d *Document) UnitByID(id string) (*Unit, bool) {
	if d == nil {
		return nil, false
	}
	u, ok := d.units[id]
	return u, ok
}

// ContextOf returns the context referenced by the given fact, if available.
func (d *Document) ContextOf(f *Fact) (*Context, bool) {
	if d == nil || f == nil {
		return nil, false
	}
	return d.ContextByID(f.ContextRef())
}

// UnitOf returns the unit referenced by the given fact, if available.
func (d *Document) UnitOf(f *Fact) (*Unit, bool) {
	if d == nil || f == nil {
		return nil, false
	}
	return d.UnitByID(f.UnitRef())
}

// Href returns the href of the schema reference.
func (s SchemaRef) Href() string {
	return s.href
}

// ID returns the context ID.
func (c *Context) ID() string {
	if c == nil {
		return ""
	}
	return c.id
}

// Entity returns the entity of the context.
func (c *Context) Entity() Entity {
	if c == nil {
		return Entity{}
	}
	return c.entity
}

// Period returns the period of the context.
func (c *Context) Period() Period {
	if c == nil {
		return Period{}
	}
	return c.period
}

// Dimensions returns a copy of the dimensions (from segment/scenario)
// associated with the context.
func (c *Context) Dimensions() []Dimension {
	if c == nil {
		return nil
	}
	out := make([]Dimension, len(c.dimensions))
	copy(out, c.dimensions)
	return out
}

// DimensionByQName returns the first dimension whose QName (URI+local)
// matches the given QName. Prefix is ignored for comparison.
func (c *Context) DimensionByQName(dim QName) (Dimension, bool) {
	if c == nil {
		return Dimension{}, false
	}
	for _, d := range c.dimensions {
		if d.dimension.uri == dim.uri && d.dimension.local == dim.local {
			return d, true
		}
	}
	return Dimension{}, false
}

// Identifier returns the identifier of the entity.
func (e Entity) Identifier() ContextIdentifier {
	return e.identifier
}

// Scheme returns the identifier scheme.
func (ci ContextIdentifier) Scheme() string {
	return ci.scheme
}

// Value returns the identifier value.
func (ci ContextIdentifier) Value() string {
	return ci.value
}

// Instant returns the instant date if the period is an instant.
func (p Period) Instant() (string, bool) {
	if p.instant == nil {
		return "", false
	}
	return *p.instant, true
}

// StartDate returns the start date of a duration period.
func (p Period) StartDate() (string, bool) {
	if p.startDate == nil {
		return "", false
	}
	return *p.startDate, true
}

// EndDate returns the end date of a duration period.
func (p Period) EndDate() (string, bool) {
	if p.endDate == nil {
		return "", false
	}
	return *p.endDate, true
}

// IsInstant reports whether the period represents an instant.
func (p Period) IsInstant() bool {
	return p.instant != nil && p.startDate == nil && p.endDate == nil && !p.forever
}

// IsForever reports whether the period represents "forever".
func (p Period) IsForever() bool {
	return p.forever
}

// ID returns the unit ID.
func (u *Unit) ID() string {
	if u == nil {
		return ""
	}
	return u.id
}

// Measures returns a copy of the simple measures of the unit.
//
// For divide units, this slice is typically empty; use
// NumeratorMeasures/DenominatorMeasures instead.
func (u *Unit) Measures() []QName {
	if u == nil {
		return nil
	}
	out := make([]QName, len(u.measures))
	copy(out, u.measures)
	return out
}

// IsDivide reports whether this unit uses a <divide> structure.
func (u *Unit) IsDivide() bool {
	if u == nil {
		return false
	}
	return u.divide
}

// NumeratorMeasures returns a copy of the measures in <unitNumerator>.
func (u *Unit) NumeratorMeasures() []QName {
	if u == nil {
		return nil
	}
	out := make([]QName, len(u.numerator))
	copy(out, u.numerator)
	return out
}

// DenominatorMeasures returns a copy of the measures in <unitDenominator>.
func (u *Unit) DenominatorMeasures() []QName {
	if u == nil {
		return nil
	}
	out := make([]QName, len(u.denominator))
	copy(out, u.denominator)
	return out
}

// Prefix returns the namespace prefix of the QName.
func (q QName) Prefix() string {
	return q.prefix
}

// Local returns the local part of the QName.
func (q QName) Local() string {
	return q.local
}

// URI returns the namespace URI of the QName.
func (q QName) URI() string {
	return q.uri
}

// String returns a string representation of the QName.
func (q QName) String() string {
	if q.uri == "" {
		if q.prefix == "" {
			return q.local
		}
		return q.prefix + ":" + q.local
	}
	// Curly-brace style
	return "{" + q.uri + "}" + q.local
}

// Kind returns the kind of the fact.
func (f *Fact) Kind() FactKind {
	if f == nil {
		return FactKindUnknown
	}
	return f.kind
}

// Name returns the QName of the fact.
func (f *Fact) Name() QName {
	if f == nil {
		return QName{}
	}
	return f.name
}

// Value returns the raw value of the fact as stored in the instance document.
func (f *Fact) Value() string {
	if f == nil {
		return ""
	}
	return f.value
}

// NormalizedValue returns a normalized form of the fact value where
// various space-like characters are converted to ASCII space and
// consecutive whitespace is collapsed into a single space.
//
// This is intended for human-readable output. The original raw value
// can be obtained via Value().
func (f *Fact) NormalizedValue() string {
	if f == nil {
		return ""
	}
	return normalizeSpace(f.value)
}

// ContextRef returns the ID of the context referenced by the fact.
func (f *Fact) ContextRef() string {
	if f == nil {
		return ""
	}
	return f.contextRef
}

// UnitRef returns the ID of the unit referenced by the fact.
func (f *Fact) UnitRef() string {
	if f == nil {
		return ""
	}
	return f.unitRef
}

// Decimals returns the decimals attribute of the fact.
func (f *Fact) Decimals() string {
	if f == nil {
		return ""
	}
	return f.decimals
}

// Precision returns the precision attribute of the fact.
func (f *Fact) Precision() string {
	if f == nil {
		return ""
	}
	return f.precision
}

// ID returns the ID attribute of the fact.
func (f *Fact) ID() string {
	if f == nil {
		return ""
	}
	return f.id
}

// Lang returns the xml:lang of the fact.
func (f *Fact) Lang() string {
	if f == nil {
		return ""
	}
	return f.lang
}

// IsNil reports whether the fact is marked as xsi:nil="true".
func (f *Fact) IsNil() bool {
	if f == nil {
		return false
	}
	return f.nil
}

// Concept represents a taxonomy concept (typically defined by xs:element
// in an XBRL schema).
type Concept struct {
	qname QName

	id string

	substitutionGroup QName
	typeName          QName

	abstract   bool
	nillable   bool
	periodType string // "instant" / "duration" / "forever" or empty
	balance    string // "debit" / "credit" or empty
}

// QName returns the QName of the concept.
func (c *Concept) QName() QName {
	if c == nil {
		return QName{}
	}
	return c.qname
}

// ID returns the @id of the concept, if any.
func (c *Concept) ID() string {
	if c == nil {
		return ""
	}
	return c.id
}

// SubstitutionGroup returns the substitutionGroup of the concept
// (e.g. xbrli:item, xbrli:tuple).
func (c *Concept) SubstitutionGroup() QName {
	if c == nil {
		return QName{}
	}
	return c.substitutionGroup
}

// Type returns the @type of the concept.
func (c *Concept) Type() QName {
	if c == nil {
		return QName{}
	}
	return c.typeName
}

// Abstract reports whether the concept is abstract.
func (c *Concept) Abstract() bool {
	if c == nil {
		return false
	}
	return c.abstract
}

// Nillable reports whether the concept is nillable.
func (c *Concept) Nillable() bool {
	if c == nil {
		return false
	}
	return c.nillable
}

// PeriodType returns the periodType ("instant"/"duration"/"forever") if set.
func (c *Concept) PeriodType() string {
	if c == nil {
		return ""
	}
	return c.periodType
}

// Balance returns the balance ("debit"/"credit") if set.
func (c *Concept) Balance() string {
	if c == nil {
		return ""
	}
	return c.balance
}

func (c *Concept) IsItem() bool {
	if c == nil {
		return false
	}
	sg := c.SubstitutionGroup()
	return sg.URI() == "http://www.xbrl.org/2003/instance" && sg.Local() == "item"
}

func (c *Concept) IsTuple() bool {
	if c == nil {
		return false
	}
	sg := c.SubstitutionGroup()
	return sg.URI() == "http://www.xbrl.org/2003/instance" && sg.Local() == "tuple"
}

// Taxonomy represents a collection of concepts from one or more schemas.
type Taxonomy struct {
	concepts map[QName]*Concept
}

// NewTaxonomy creates an empty taxonomy.
func NewTaxonomy() *Taxonomy {
	return &Taxonomy{
		concepts: make(map[QName]*Concept),
	}
}

// Concepts returns a copy of the concepts map (QName -> *Concept).
func (t *Taxonomy) Concepts() map[QName]*Concept {
	if t == nil {
		return nil
	}
	out := make(map[QName]*Concept, len(t.concepts))
	for k, v := range t.concepts {
		out[k] = v
	}
	return out
}

// Concept returns the concept for the given QName, if present.
func (t *Taxonomy) Concept(q QName) (*Concept, bool) {
	if t == nil {
		return nil, false
	}
	c, ok := t.concepts[q]
	return c, ok
}

// addConcept inserts or replaces a concept in the taxonomy.
// (internal; used by the taxonomy parser)
func (t *Taxonomy) addConcept(c *Concept) {
	if t == nil || c == nil {
		return
	}
	if t.concepts == nil {
		t.concepts = make(map[QName]*Concept)
	}
	t.concepts[c.qname] = c
}

// Taxonomy returns the taxonomy attached to the document, if any.
func (d *Document) Taxonomy() *Taxonomy {
	if d == nil {
		return nil
	}
	return d.taxonomy
}

// SetTaxonomy attaches the given taxonomy to the document.
func (d *Document) SetTaxonomy(t *Taxonomy) {
	if d == nil {
		return
	}
	d.taxonomy = t
}

// LoadTaxonomyFromSchemaRefs builds a Taxonomy from this Document's
// schemaRefs using the provided opener, and attaches it to the Document.
func (d *Document) LoadTaxonomyFromSchemaRefs(
	opener func(href string) (io.ReadCloser, error),
) (*Taxonomy, error) {
	if d == nil {
		return nil, fmt.Errorf("xbrl: document is nil")
	}
	if opener == nil {
		return nil, fmt.Errorf("xbrl: opener is nil")
	}

	tax := NewTaxonomy()

	for _, sr := range d.schemaRefs {
		href := sr.Href()
		if href == "" {
			continue
		}

		rc, err := opener(href)
		if err != nil {
			return nil, fmt.Errorf("xbrl: open schemaRef %q: %w", href, err)
		}

		t, err := ParseTaxonomy(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("xbrl: parse schemaRef %q: %w", href, err)
		}

		tax.Merge(t)
	}

	d.taxonomy = tax
	return tax, nil
}

// ConceptOf returns the taxonomy concept corresponding to the fact's
// QName, if a taxonomy is attached and the concept exists.
func (d *Document) ConceptOf(f *Fact) (*Concept, bool) {
	if d == nil || f == nil || d.taxonomy == nil {
		return nil, false
	}
	return d.taxonomy.Concept(f.Name())
}
