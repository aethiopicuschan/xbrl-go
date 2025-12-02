package xbrl

import (
	"encoding/json"
	"io"
)

// FactJSON is a simple DTO for exporting facts as JSON.
type FactJSON struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	ContextRef string `json:"context"`
	UnitRef    string `json:"unit"`
	Nil        bool   `json:"nil"`
}

// FactsAsJSONDTOs converts all facts in a Document into a slice of
// FactJSON DTOs.
func (d *Document) FactsAsJSONDTOs() []FactJSON {
	if d == nil {
		return nil
	}
	out := make([]FactJSON, 0, len(d.facts))
	for _, f := range d.facts {
		if f == nil {
			continue
		}
		value := f.Value()
		if f.IsNil() {
			value = ""
		}
		out = append(out, FactJSON{
			Name:       f.Name().String(),
			Value:      value,
			ContextRef: f.ContextRef(),
			UnitRef:    f.UnitRef(),
			Nil:        f.IsNil(),
		})
	}
	return out
}

// EncodeFactsJSON writes all facts in the Document as JSON array to w.
// - HTML escape is disabled
// - If pretty is true, indented output is used
func (d *Document) EncodeFactsJSON(w io.Writer, pretty bool) error {
	if d == nil {
		return nil
	}

	enc := json.NewEncoder(w)
	if pretty {
		enc.SetIndent("", "  ")
	}
	enc.SetEscapeHTML(false)

	dtos := d.FactsAsJSONDTOs()
	return enc.Encode(dtos)
}
