package xbrl

import (
	"encoding/xml"
	"fmt"
	"io"
	"maps"
	"os"
	"strings"
)

// ParseFile parses an XBRL instance document from a file path.
func ParseFile(path string) (*Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("xbrl: open file: %w", err)
	}
	defer f.Close()

	return Parse(f)
}

// Parse parses an XBRL instance document from an io.Reader.
func Parse(r io.Reader) (*Document, error) {
	dec := xml.NewDecoder(r)
	dec.CharsetReader = charsetReader

	var doc Document
	doc.contexts = make(map[string]*Context)
	doc.units = make(map[string]*Unit)

	nsMap := newNamespaceStack()

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("xbrl: decode token: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			nsMap.Push(t)

			if isXbrlRoot(t) {
				continue
			}

			switch {
			case isSchemaRef(t):
				sr := parseSchemaRef(t)
				doc.schemaRefs = append(doc.schemaRefs, sr)

			case t.Name.Local == "context":
				ctx, err := parseContext(dec, t, nsMap)
				if err != nil {
					return nil, err
				}
				doc.contexts[ctx.id] = ctx

			case t.Name.Local == "unit":
				unit, err := parseUnit(dec, t, nsMap)
				if err != nil {
					return nil, err
				}
				doc.units[unit.id] = unit

			default:
				// item facts (simplified detection)
				if hasAttr(t.Attr, "contextRef") {
					fact, err := parseItemFact(dec, t, nsMap)
					if err != nil {
						return nil, err
					}
					doc.facts = append(doc.facts, fact)
				}
			}

		case xml.EndElement:
			nsMap.Pop(t)
		}
	}

	return &doc, nil
}

// ---------- Element detection / small parsers ----------

func isXbrlRoot(se xml.StartElement) bool {
	// XBRL root element is usually "xbrl"
	return strings.EqualFold(se.Name.Local, "xbrl")
}

func isSchemaRef(se xml.StartElement) bool {
	return se.Name.Local == "schemaRef"
}

func parseSchemaRef(se xml.StartElement) SchemaRef {
	var href string
	for _, a := range se.Attr {
		if a.Name.Local == "href" {
			href = a.Value
			break
		}
	}
	return SchemaRef{href: href}
}

func parseContext(dec *xml.Decoder, start xml.StartElement, ns *namespaceStack) (*Context, error) {
	ctx := &Context{}
	for _, a := range start.Attr {
		if a.Name.Local == "id" {
			ctx.id = a.Value
		}
	}

	var dims []Dimension

	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("xbrl: parse context: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "entity":
				ent, segDims, err := parseEntity(dec, t, ns)
				if err != nil {
					return nil, err
				}
				ctx.entity = *ent
				dims = append(dims, segDims...)
			case "period":
				p, err := parsePeriod(dec, t)
				if err != nil {
					return nil, err
				}
				ctx.period = *p
			case "scenario":
				scnDims, err := parseDimensionsContainer(dec, t, ns)
				if err != nil {
					return nil, err
				}
				dims = append(dims, scnDims...)
			default:
				if err := dec.Skip(); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				ctx.dimensions = dims
				return ctx, nil
			}
		}
	}
}

func parseEntity(dec *xml.Decoder, start xml.StartElement, ns *namespaceStack) (*Entity, []Dimension, error) {
	ent := &Entity{}
	var dims []Dimension

	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, nil, fmt.Errorf("xbrl: parse entity: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "identifier":
				var ident ContextIdentifier
				for _, a := range t.Attr {
					if a.Name.Local == "scheme" {
						ident.scheme = a.Value
					}
				}
				var value string
				if err := dec.DecodeElement(&value, &t); err != nil {
					return nil, nil, fmt.Errorf("xbrl: parse identifier: %w", err)
				}
				ident.value = strings.TrimSpace(value)
				ent.identifier = ident
			case "segment":
				segDims, err := parseDimensionsContainer(dec, t, ns)
				if err != nil {
					return nil, nil, err
				}
				dims = append(dims, segDims...)
			default:
				if err := dec.Skip(); err != nil {
					return nil, nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return ent, dims, nil
			}
		}
	}
}

func parsePeriod(dec *xml.Decoder, start xml.StartElement) (*Period, error) {
	p := &Period{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("xbrl: parse period: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "instant":
				var v string
				if err := dec.DecodeElement(&v, &t); err != nil {
					return nil, err
				}
				v = strings.TrimSpace(v)
				p.instant = &v
			case "startDate":
				var v string
				if err := dec.DecodeElement(&v, &t); err != nil {
					return nil, err
				}
				v = strings.TrimSpace(v)
				p.startDate = &v
			case "endDate":
				var v string
				if err := dec.DecodeElement(&v, &t); err != nil {
					return nil, err
				}
				v = strings.TrimSpace(v)
				p.endDate = &v
			case "forever":
				if err := dec.Skip(); err != nil {
					return nil, err
				}
				p.forever = true
			default:
				if err := dec.Skip(); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return p, nil
			}
		}
	}
}

func parseUnit(dec *xml.Decoder, start xml.StartElement, ns *namespaceStack) (*Unit, error) {
	u := &Unit{}
	for _, a := range start.Attr {
		if a.Name.Local == "id" {
			u.id = a.Value
		}
	}

	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("xbrl: parse unit: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "measure":
				// simple unit measure (top-level)
				q, err := parseMeasureElement(dec, t, ns)
				if err != nil {
					return nil, err
				}
				u.measures = append(u.measures, q)
			case "divide":
				// divide unit
				num, den, err := parseDivide(dec, t, ns)
				if err != nil {
					return nil, err
				}
				u.divide = true
				u.numerator = num
				u.denominator = den
			default:
				if err := dec.Skip(); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return u, nil
			}
		}
	}
}

// parseMeasureElement parses a <measure> element into a QName.
func parseMeasureElement(dec *xml.Decoder, start xml.StartElement, ns *namespaceStack) (QName, error) {
	var v string
	if err := dec.DecodeElement(&v, &start); err != nil {
		return QName{}, err
	}
	v = strings.TrimSpace(v)

	prefix := prefixOf(v)
	local := localOf(v)
	uri := ""
	if ns != nil {
		if prefix == "" {
			uri = ns.URIForPrefix("")
		} else {
			uri = ns.URIForPrefix(prefix)
		}
	}

	return QName{
		prefix: prefix,
		local:  local,
		uri:    uri,
	}, nil
}

// parseDivide parses a <divide> element into numerator/denominator measures.
func parseDivide(dec *xml.Decoder, start xml.StartElement, ns *namespaceStack) ([]QName, []QName, error) {
	var num, den []QName

	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, nil, fmt.Errorf("xbrl: parse divide: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "unitNumerator":
				n, err := parseUnitMeasureContainer(dec, t, ns)
				if err != nil {
					return nil, nil, err
				}
				num = n
			case "unitDenominator":
				d, err := parseUnitMeasureContainer(dec, t, ns)
				if err != nil {
					return nil, nil, err
				}
				den = d
			default:
				if err := dec.Skip(); err != nil {
					return nil, nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return num, den, nil
			}
		}
	}
}

// parseUnitMeasureContainer parses <unitNumerator> or <unitDenominator>.
func parseUnitMeasureContainer(dec *xml.Decoder, start xml.StartElement, ns *namespaceStack) ([]QName, error) {
	var measures []QName

	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("xbrl: parse unit measure container: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "measure" {
				q, err := parseMeasureElement(dec, t, ns)
				if err != nil {
					return nil, err
				}
				measures = append(measures, q)
			} else {
				if err := dec.Skip(); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return measures, nil
			}
		}
	}
}

func parseItemFact(dec *xml.Decoder, start xml.StartElement, ns *namespaceStack) (*Fact, error) {
	prefix := ""
	if ns != nil {
		prefix = ns.PrefixForURI(start.Name.Space)
	}

	f := &Fact{
		kind: FactKindItem,
		name: QName{
			prefix: prefix,
			local:  start.Name.Local,
			uri:    start.Name.Space,
		},
	}

	for _, a := range start.Attr {
		switch a.Name.Local {
		case "contextRef":
			f.contextRef = a.Value
		case "unitRef":
			f.unitRef = a.Value
		case "decimals":
			f.decimals = a.Value
		case "precision":
			f.precision = a.Value
		case "id":
			f.id = a.Value
		case "lang":
			f.lang = a.Value
		}

		// xsi:nil="true"
		if a.Name.Space == "http://www.w3.org/2001/XMLSchema-instance" && a.Name.Local == "nil" {
			if strings.EqualFold(a.Value, "true") {
				f.nil = true
			}
		}
	}

	var value string
	if err := dec.DecodeElement(&value, &start); err != nil {
		return nil, fmt.Errorf("xbrl: parse fact %s: %w", start.Name.Local, err)
	}
	f.value = strings.TrimSpace(value)

	return f, nil
}

// parseDimensionsContainer parses a <segment> or <scenario> element and
// returns all explicit/typed dimensions contained within it.
func parseDimensionsContainer(dec *xml.Decoder, start xml.StartElement, ns *namespaceStack) ([]Dimension, error) {
	var dims []Dimension

	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("xbrl: parse dimensions (%s): %w", start.Name.Local, err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "explicitMember":
				d, err := parseExplicitMember(dec, t, ns)
				if err != nil {
					return nil, err
				}
				dims = append(dims, d)
			case "typedMember":
				d, err := parseTypedMember(dec, t, ns)
				if err != nil {
					return nil, err
				}
				dims = append(dims, d)
			default:
				if err := dec.Skip(); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return dims, nil
			}
		}
	}
}

func parseExplicitMember(dec *xml.Decoder, start xml.StartElement, ns *namespaceStack) (Dimension, error) {
	var dimName string
	for _, a := range start.Attr {
		if a.Name.Local == "dimension" {
			dimName = strings.TrimSpace(a.Value)
			break
		}
	}
	// dimension QName
	dimPrefix := prefixOf(dimName)
	dimLocal := localOf(dimName)
	dimURI := ""
	if ns != nil {
		if dimPrefix == "" {
			dimURI = ns.URIForPrefix("")
		} else {
			dimURI = ns.URIForPrefix(dimPrefix)
		}
	}
	dimQ := QName{
		prefix: dimPrefix,
		local:  dimLocal,
		uri:    dimURI,
	}

	// member QName from element text
	var value string
	if err := dec.DecodeElement(&value, &start); err != nil {
		return Dimension{}, fmt.Errorf("xbrl: parse explicitMember: %w", err)
	}
	value = strings.TrimSpace(value)
	memPrefix := prefixOf(value)
	memLocal := localOf(value)
	memURI := ""
	if ns != nil {
		if memPrefix == "" {
			memURI = ns.URIForPrefix("")
		} else {
			memURI = ns.URIForPrefix(memPrefix)
		}
	}
	memQ := QName{
		prefix: memPrefix,
		local:  memLocal,
		uri:    memURI,
	}

	return Dimension{
		dimension:  dimQ,
		explicit:   true,
		member:     memQ,
		typedValue: "",
	}, nil
}

func parseTypedMember(dec *xml.Decoder, start xml.StartElement, ns *namespaceStack) (Dimension, error) {
	_ = ns // currently unused, but kept for symmetry/future use

	var dimName string
	for _, a := range start.Attr {
		if a.Name.Local == "dimension" {
			dimName = strings.TrimSpace(a.Value)
			break
		}
	}
	// dimension QName
	dimPrefix := prefixOf(dimName)
	dimLocal := localOf(dimName)
	dimURI := ""
	if ns != nil {
		if dimPrefix == "" {
			dimURI = ns.URIForPrefix("")
		} else {
			dimURI = ns.URIForPrefix(dimPrefix)
		}
	}
	dimQ := QName{
		prefix: dimPrefix,
		local:  dimLocal,
		uri:    dimURI,
	}

	// grab inner XML as-is
	type inner struct {
		XML string `xml:",innerxml"`
	}
	var in inner
	if err := dec.DecodeElement(&in, &start); err != nil {
		return Dimension{}, fmt.Errorf("xbrl: parse typedMember: %w", err)
	}

	return Dimension{
		dimension:  dimQ,
		explicit:   false,
		member:     QName{},
		typedValue: strings.TrimSpace(in.XML),
	}, nil
}

// ---------- small utilities ----------

func hasAttr(attrs []xml.Attr, local string) bool {
	for _, a := range attrs {
		if a.Name.Local == local {
			return true
		}
	}
	return false
}

func prefixOf(s string) string {
	i := strings.IndexByte(s, ':')
	if i < 0 {
		return ""
	}
	return s[:i]
}

func localOf(s string) string {
	i := strings.IndexByte(s, ':')
	if i < 0 {
		return s
	}
	return s[i+1:]
}

// charsetReader is a placeholder. For now, we assume UTF-8 only.
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	// TODO : implement charset decoding if needed
	return input, nil
}

// ---------- namespace stack (for URI resolution) ----------

type namespaceStack struct {
	stack []map[string]string // prefix -> URI
}

func newNamespaceStack() *namespaceStack {
	return &namespaceStack{
		stack: []map[string]string{{}},
	}
}

// Push adds a new namespace context to the stack based on the given start element.
func (ns *namespaceStack) Push(se xml.StartElement) {
	top := map[string]string{}
	if len(ns.stack) > 0 {
		maps.Copy(top, ns.stack[len(ns.stack)-1])
	}

	for _, a := range se.Attr {
		if a.Name.Space == "xmlns" {
			// xmlns:prefix="URI"
			top[a.Name.Local] = a.Value
		} else if a.Name.Local == "xmlns" && a.Name.Space == "" {
			// default namespace: xmlns="URI"
			top[""] = a.Value
		}
	}

	ns.stack = append(ns.stack, top)
}

// Pop removes the top namespace context from the stack.
func (ns *namespaceStack) Pop(_ xml.EndElement) {
	if len(ns.stack) > 1 {
		ns.stack = ns.stack[:len(ns.stack)-1]
	}
}

// URIForPrefix returns the namespace URI for the given prefix in the current namespace context.
func (ns *namespaceStack) URIForPrefix(prefix string) string {
	if len(ns.stack) == 0 {
		return ""
	}
	top := ns.stack[len(ns.stack)-1]
	return top[prefix]
}

// PrefixForURI returns the first prefix found for the given URI in the current namespace context.
func (ns *namespaceStack) PrefixForURI(uri string) string {
	if len(ns.stack) == 0 || uri == "" {
		return ""
	}
	top := ns.stack[len(ns.stack)-1]
	for p, u := range top {
		if u == uri {
			return p
		}
	}
	return ""
}
