package logseq

import (
	"encoding/json"
	"strings"
)

// Property represents a property of a block in a Logseq graph.
type Property struct {
	Name  string
	Value string
}

// String returns the value of the property as a string.
// If the value is a page link, the link syntax is removed.
func (p *Property) String() string {
	if p.IsPageLink() {
		return strings.TrimSuffix(strings.TrimPrefix(p.Value, "[["), "]]")
	}
	return p.Value
}

// Bool returns the value of the property interpreted as a boolean.
func (p *Property) Bool() bool {
	return p.Value == "true"
}

// List returns the value of the property as a list of strings.
// The value is split by commas and spaces.
func (p *Property) List() []string {
	return strings.Split(p.Value, ", ")
}

// IsPageLink returns true if the value of the property is a page link.
func (p *Property) IsPageLink() bool {
	return strings.HasPrefix(p.Value, "[[") && strings.HasSuffix(p.Value, "]]")
}

type PropertyMap struct {
	Properties map[string]*Property
}

func NewPropertyMap() *PropertyMap {
	return &PropertyMap{
		Properties: map[string]*Property{},
	}
}

func (pm *PropertyMap) Get(name string) (*Property, bool) {
	if pm.Properties[name] == nil {
		return nil, false
	}

	return pm.Properties[name], true
}

func (pm *PropertyMap) Set(name string, value string) {
	pm.Properties[name] = &Property{
		Name:  name,
		Value: value,
	}
}

func (pm *PropertyMap) MarshalJSON() ([]byte, error) {
	propsMap := map[string]string{}
	for name, prop := range pm.Properties {
		propsMap[name] = prop.String()
	}

	return json.Marshal(&propsMap)
}
