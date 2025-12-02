package xbrl

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// ParseTaxonomyFile parses an XBRL taxonomy schema (XSD) from a file path.
func ParseTaxonomyFile(path string) (*Taxonomy, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("xbrl: open taxonomy schema: %w", err)
	}
	defer f.Close()
	return ParseTaxonomy(f)
}

// ParseTaxonomy parses an XBRL taxonomy schema (XSD) from an io.Reader.
//
// This function focuses on xs:element declarations and extracts basic
// concept information such as name, id, substitutionGroup, type,
// abstract, nillable, periodType, and balance.
//
// It is intentionally minimal and does not attempt to parse linkbases
// (labels, presentation, calculation, etc.).
func ParseTaxonomy(r io.Reader) (*Taxonomy, error) {
	dec := xml.NewDecoder(r)

	ns := newNamespaceStack()
	tax := NewTaxonomy()

	var targetNS string

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("xbrl: decode taxonomy token: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			ns.Push(t)

			switch t.Name.Local {
			case "schema":
				for _, a := range t.Attr {
					if a.Name.Local == "targetNamespace" {
						targetNS = strings.TrimSpace(a.Value)
						break
					}
				}

			case "element":
				c := conceptFromElement(t, targetNS, ns)
				if c != nil {
					tax.addConcept(c)
				}
				// skip element contents (annotation, etc.)
				if err := dec.Skip(); err != nil {
					return nil, fmt.Errorf("xbrl: skip element: %w", err)
				}
			}

		case xml.EndElement:
			ns.Pop(t)
		}
	}

	return tax, nil
}

// conceptFromElement creates a Concept from an xs:element start tag.
//
// It only looks at attributes and does not consume any child tokens.
func conceptFromElement(se xml.StartElement, targetNS string, ns *namespaceStack) *Concept {
	var (
		name  string
		id    string
		typ   string
		subst string

		abstractStr string
		nillableStr string
		periodType  string
		balance     string
	)

	for _, a := range se.Attr {
		switch a.Name.Local {
		case "name":
			name = strings.TrimSpace(a.Value)
		case "id":
			id = strings.TrimSpace(a.Value)
		case "type":
			typ = strings.TrimSpace(a.Value)
		case "substitutionGroup":
			subst = strings.TrimSpace(a.Value)
		case "abstract":
			abstractStr = strings.TrimSpace(a.Value)
		case "nillable":
			nillableStr = strings.TrimSpace(a.Value)
		case "periodType":
			periodType = strings.TrimSpace(a.Value)
		case "balance":
			balance = strings.TrimSpace(a.Value)
		}
	}

	if name == "" || targetNS == "" {
		// Without a name or target namespace we cannot form a proper concept QName.
		return nil
	}

	// Concept QName is (targetNamespace, name).
	conceptPrefix := ""
	if ns != nil {
		conceptPrefix = ns.PrefixForURI(targetNS)
	}
	cq := QName{
		prefix: conceptPrefix,
		local:  name,
		uri:    targetNS,
	}

	// substitutionGroup QName (e.g. xbrli:item).
	var sgQName QName
	if subst != "" {
		p := prefixOf(subst)
		l := localOf(subst)
		u := ""
		if ns != nil {
			if p == "" {
				u = ns.URIForPrefix("")
			} else {
				u = ns.URIForPrefix(p)
			}
		}
		sgQName = QName{
			prefix: p,
			local:  l,
			uri:    u,
		}
	}

	// type QName (e.g. xbrli:monetaryItemType).
	var typeQName QName
	if typ != "" {
		p := prefixOf(typ)
		l := localOf(typ)
		u := ""
		if ns != nil {
			if p == "" {
				u = ns.URIForPrefix("")
			} else {
				u = ns.URIForPrefix(p)
			}
		}
		typeQName = QName{
			prefix: p,
			local:  l,
			uri:    u,
		}
	}

	c := &Concept{
		qname:             cq,
		id:                id,
		substitutionGroup: sgQName,
		typeName:          typeQName,
		abstract:          parseBool(abstractStr),
		nillable:          parseBool(nillableStr),
		periodType:        periodType,
		balance:           balance,
	}

	return c
}

// Merge merges concepts from other into t.
// Existing concepts with the same QName are overwritten.
func (t *Taxonomy) Merge(other *Taxonomy) {
	if t == nil || other == nil {
		return
	}
	if t.concepts == nil {
		t.concepts = make(map[QName]*Concept)
	}
	for q, c := range other.concepts {
		t.concepts[q] = c
	}
}

// parseBool interprets common boolean lexical forms.
// Only "true" / "1" (case-insensitive) are treated as true.
func parseBool(s string) bool {
	if s == "" {
		return false
	}
	switch strings.ToLower(s) {
	case "true", "1":
		return true
	default:
		return false
	}
}
