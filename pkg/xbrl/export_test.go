package xbrl

// NOTE: Test-only helper constructors to access unexported fields.
// This file is compiled only in tests.

func NewSchemaRefForTest(href string) SchemaRef {
	return SchemaRef{href: href}
}

func NewContextIdentifierForTest(scheme, value string) ContextIdentifier {
	return ContextIdentifier{
		scheme: scheme,
		value:  value,
	}
}

func NewEntityForTest(id ContextIdentifier) Entity {
	return Entity{
		identifier: id,
	}
}

func NewPeriodForTest(instant, start, end *string, forever bool) Period {
	return Period{
		instant:   instant,
		startDate: start,
		endDate:   end,
		forever:   forever,
	}
}

func NewQNameForTest(prefix, local, uri string) QName {
	return QName{
		prefix: prefix,
		local:  local,
		uri:    uri,
	}
}

func NewDimensionForTest(dim QName, explicit bool, member QName, typedValue string) Dimension {
	return Dimension{
		dimension:  dim,
		explicit:   explicit,
		member:     member,
		typedValue: typedValue,
	}
}

func NewContextForTest(id string, entity Entity, period Period, dims []Dimension) *Context {
	return &Context{
		id:         id,
		entity:     entity,
		period:     period,
		dimensions: dims,
	}
}

func NewUnitSimpleForTest(id string, measures []QName) *Unit {
	return &Unit{
		id:       id,
		measures: measures,
	}
}

func NewUnitDivideForTest(id string, numerator, denominator []QName) *Unit {
	return &Unit{
		id:          id,
		divide:      true,
		numerator:   numerator,
		denominator: denominator,
	}
}

func NewConceptForTest(
	q QName,
	id string,
	subst QName,
	typ QName,
	abstract bool,
	nillable bool,
	periodType string,
	balance string,
) *Concept {
	return &Concept{
		qname:             q,
		id:                id,
		substitutionGroup: subst,
		typeName:          typ,
		abstract:          abstract,
		nillable:          nillable,
		periodType:        periodType,
		balance:           balance,
	}
}

func NewTaxonomyForTest(concepts map[QName]*Concept) *Taxonomy {
	return &Taxonomy{
		concepts: concepts,
	}
}

func NewFactForTest(
	kind FactKind,
	name QName,
	value string,
	contextRef string,
	unitRef string,
	decimals string,
	precision string,
	id string,
	lang string,
	isNil bool,
) *Fact {
	return &Fact{
		kind:       kind,
		name:       name,
		value:      value,
		contextRef: contextRef,
		unitRef:    unitRef,
		decimals:   decimals,
		precision:  precision,
		id:         id,
		lang:       lang,
		nil:        isNil,
	}
}

func NewDocumentForTest(
	schemaRefs []SchemaRef,
	contexts map[string]*Context,
	units map[string]*Unit,
	facts []*Fact,
	tax *Taxonomy,
) *Document {
	return &Document{
		schemaRefs: schemaRefs,
		contexts:   contexts,
		units:      units,
		facts:      facts,
		taxonomy:   tax,
	}
}

var NormalizeSpace = normalizeSpace
