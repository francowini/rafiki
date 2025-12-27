package telegrammessage

import (
	"testing"
)

func TestMergePhysicalData(t *testing.T) {
	tests := []struct {
		name      string
		sintomas  string
		emociones string
		want      string
	}{
		{"both_present", "palpitaciones", "ansiedad", "palpitaciones | ansiedad"},
		{"only_sintomas", "palpitaciones", "", "palpitaciones"},
		{"only_emociones", "", "ansiedad", "ansiedad"},
		{"both_empty", "", "", ""},
		{"whitespace", "  palpitaciones  ", "  ansiedad  ", "palpitaciones | ansiedad"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergePhysicalData(tt.sintomas, tt.emociones)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseIntensity(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    int
		wantErr bool
	}{
		{"valid_5", "5", 5, false},
		{"valid_0", "0", 0, false},
		{"valid_10", "10", 10, false},
		{"whitespace", "  7  ", 7, false},
		{"empty", "", 0, true},
		{"invalid_11", "11", 0, true},
		{"invalid_text", "alto", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIntensity(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Value() != tt.want {
				t.Errorf("got %v, want %v", got.Value(), tt.want)
			}
		})
	}
}
