package xbrl

// FactFilter describes criteria to filter facts.
//
// All fields are unexported and should be configured via the builder-style
// methods (ConceptURI, ConceptLocal, ContextID, UnitID, OnlyNil, ExcludeNil, Dimension).
type FactFilter struct {
	conceptURI   string
	conceptLocal string
	contextID    string
	unitID       string
	nilFilter    *bool

	// dims holds required explicit dimensions.
	// A fact matches only if its context has *all* of these
	// dimension/member pairs as explicit dimensions.
	dims []dimensionFilter
}

// dimensionFilter describes one explicit dimension requirement.
type dimensionFilter struct {
	dimURI, dimLocal string
	memURI, memLocal string
}

// NewFactFilter creates an empty fact filter.
func NewFactFilter() *FactFilter {
	return &FactFilter{}
}

// ConceptURI sets the expected namespace URI for the fact concept.
func (f *FactFilter) ConceptURI(uri string) *FactFilter {
	if f == nil {
		return nil
	}
	f.conceptURI = uri
	return f
}

// ConceptLocal sets the expected local name for the fact concept.
func (f *FactFilter) ConceptLocal(local string) *FactFilter {
	if f == nil {
		return nil
	}
	f.conceptLocal = local
	return f
}

// ContextID sets the expected context ID for the fact.
func (f *FactFilter) ContextID(id string) *FactFilter {
	if f == nil {
		return nil
	}
	f.contextID = id
	return f
}

// UnitID sets the expected unit ID for the fact.
func (f *FactFilter) UnitID(id string) *FactFilter {
	if f == nil {
		return nil
	}
	f.unitID = id
	return f
}

// OnlyNil filters for xsi:nil="true".
func (f *FactFilter) OnlyNil() *FactFilter {
	if f == nil {
		return nil
	}
	v := true
	f.nilFilter = &v
	return f
}

// ExcludeNil filters for xsi:nil="false".
func (f *FactFilter) ExcludeNil() *FactFilter {
	if f == nil {
		return nil
	}
	v := false
	f.nilFilter = &v
	return f
}

// Dimension adds an explicit dimension requirement to the filter.
//
// A fact matches the filter only if its context contains an explicit
// dimension whose dimension QName matches dim (URI+local) and whose
// member QName matches member (URI+local).
//
// Prefixes of the given QNames are ignored for comparison.
func (f *FactFilter) Dimension(dim, member QName) *FactFilter {
	if f == nil {
		return nil
	}
	df := dimensionFilter{
		dimURI:   dim.URI(),
		dimLocal: dim.Local(),
		memURI:   member.URI(),
		memLocal: member.Local(),
	}
	f.dims = append(f.dims, df)
	return f
}

// FilterFacts returns a slice of facts that match the given filter.
//
// The returned slice is a shallow copy and can be modified by the caller
// without affecting the Document.
//
// Note: dimension filters (added via Dimension) are evaluated against
// explicit dimensions on the fact's context. Typed dimensions are
// currently ignored for filtering.
func (d *Document) FilterFacts(f *FactFilter) []*Fact {
	if d == nil || f == nil {
		return nil
	}
	var result []*Fact
	for _, fact := range d.facts {
		if fact == nil {
			continue
		}

		// Concept filter
		if f.conceptLocal != "" || f.conceptURI != "" {
			q := fact.Name()
			if f.conceptLocal != "" && q.Local() != f.conceptLocal {
				continue
			}
			if f.conceptURI != "" && q.URI() != f.conceptURI {
				continue
			}
		}

		// Context filter (by ID)
		if f.contextID != "" && fact.ContextRef() != f.contextID {
			continue
		}

		// Unit filter
		if f.unitID != "" && fact.UnitRef() != f.unitID {
			continue
		}

		// Nil filter
		if f.nilFilter != nil && fact.IsNil() != *f.nilFilter {
			continue
		}

		// Dimension filters (explicit-only for now)
		if len(f.dims) > 0 {
			ctx, ok := d.contexts[fact.ContextRef()]
			if !ok || ctx == nil {
				continue
			}
			// We can use ctx.dimensions directly here since we're in the same package.
			ctxDims := ctx.dimensions

			matchAll := true
			for _, df := range f.dims {
				found := false
				for _, cd := range ctxDims {
					if !cd.explicit {
						continue
					}
					dq := cd.dimension
					mq := cd.member
					if dq.uri == df.dimURI && dq.local == df.dimLocal &&
						mq.uri == df.memURI && mq.local == df.memLocal {
						found = true
						break
					}
				}
				if !found {
					matchAll = false
					break
				}
			}
			if !matchAll {
				continue
			}
		}

		result = append(result, fact)
	}

	out := make([]*Fact, len(result))
	copy(out, result)
	return out
}
