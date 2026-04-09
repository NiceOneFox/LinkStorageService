package generator_test

import (
	"testing"

	"LinkStorageService/internal/generator"
)

func TestBase62Encoder_Encode(t *testing.T) {
	enc := generator.NewBase62Encoder()

	tests := []struct {
		name  string
		input uint64
		want  string
	}{
		{"0", 0, "0"},
		{"1", 1, "1"},
		{"9", 9, "9"},
		{"10", 10, "A"},
		{"35", 35, "Z"},
		{"36", 36, "a"},
		{"61", 61, "z"},
		{"62", 62, "10"},
		{"63", 63, "11"},
		{"3844", 3844, "100"},
		{"18446744073709551615", 18446744073709551615, "LygHa16AHYF"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := enc.Encode(tt.input)
			if got != tt.want {
				t.Errorf("Encode(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBase62Encoder_Decode(t *testing.T) {
	enc := generator.NewBase62Encoder()

	tests := []struct {
		name  string
		input string
		want  uint64
	}{
		{"0", "0", 0},
		{"1", "1", 1},
		{"9", "9", 9},
		{"A", "A", 10},
		{"Z", "Z", 35},
		{"a", "a", 36},
		{"z", "z", 61},
		{"10", "10", 62},
		{"11", "11", 63},
		{"100", "100", 3844},
		{"max", "LygHa16AHYF", 18446744073709551615},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := enc.Decode(tt.input)
			if err != nil {
				t.Errorf("Decode(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("Decode(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
