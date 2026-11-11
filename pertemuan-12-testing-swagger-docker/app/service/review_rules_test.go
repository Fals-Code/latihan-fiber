package service

import (
	"slices"
	"testing"
)

func TestNormalizeTags(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "lowercase and trim", in: []string{" Tugas-Banyak ", "SANTAI"}, want: []string{"tugas-banyak", "santai"}},
		{name: "deduplicate normalized values", in: []string{"Seru", "seru", " SERU"}, want: []string{"seru"}},
		{name: "remove empty values", in: []string{"", "   ", "praktis"}, want: []string{"praktis"}},
		{name: "retain first occurrence order", in: []string{"b", "a", "B"}, want: []string{"b", "a"}},
		{name: "empty input", in: []string{}, want: []string{}},
		{name: "nil input", in: nil, want: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeTags(tt.in)
			if !slices.Equal(got, tt.want) {
				t.Errorf("NormalizeTags(%q) = %q, want %q", tt.in, got, tt.want)
			}
			if got == nil {
				t.Errorf("NormalizeTags(%q) returned nil", tt.in)
			}
		})
	}
}
