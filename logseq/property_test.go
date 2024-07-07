package logseq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"export-logseq/logseq"
)

func TestProperty_String(t *testing.T) {
	stringTests := []struct {
		value string
		want  string
	}{
		{"value", "value"},
		{"", ""},
		{"[[value]]", "value"},
	}

	for _, tt := range stringTests {
		prop := logseq.Property{
			Name:  "test",
			Value: tt.value,
		}
		got := prop.String()
		assert.Equal(t, tt.want, got)
	}
}

func TestProperty_Bool(t *testing.T) {
	boolTests := []struct {
		value string
		want  bool
	}{
		{"true", true},
		{"false", false},
		{"", false},
	}

	for _, tt := range boolTests {
		prop := logseq.Property{
			Name:  "test",
			Value: tt.value,
		}
		got := prop.Bool()
		assert.Equal(t, tt.want, got)
	}
}

func TestProperty_List(t *testing.T) {
	prop := logseq.Property{
		Name:  "test",
		Value: "a, b, c",
	}
	want := []string{"a", "b", "c"}
	got := prop.List()
	assert.Equal(t, want, got)
}

func TestProperty_IsPageLink(t *testing.T) {
	pageLinkTests := []struct {
		value string
		want  bool
	}{
		{"[[page]]", true},
		{"page", false},
		{"", false},
	}

	for _, tt := range pageLinkTests {
		prop := logseq.Property{
			Name:  "test",
			Value: tt.value,
		}
		got := prop.IsPageLink()
		assert.Equal(t, tt.want, got)
	}
}
