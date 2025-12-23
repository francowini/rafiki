package contribution

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		want    int
		wantErr bool
	}{
		{
			name:    "valid minimum boundary (0)",
			value:   0,
			want:    0,
			wantErr: false,
		},
		{
			name:    "valid maximum boundary (10)",
			value:   10,
			want:    10,
			wantErr: false,
		},
		{
			name:    "valid middle value (5)",
			value:   5,
			want:    5,
			wantErr: false,
		},
		{
			name:    "invalid below minimum (-1)",
			value:   -1,
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid above maximum (11)",
			value:   11,
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid large negative",
			value:   -100,
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid large positive",
			value:   100,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Value() != tt.want {
				t.Errorf("Parse() = %v, want %v", got.Value(), tt.want)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	t.Run("valid value does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustParse panicked unexpectedly: %v", r)
			}
		}()

		c := MustParse(5)
		if c.Value() != 5 {
			t.Errorf("MustParse() = %v, want 5", c.Value())
		}
	})

	t.Run("invalid value panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustParse did not panic for invalid value")
			}
		}()

		MustParse(11)
	})
}

func TestContribution_String(t *testing.T) {
	c := MustParse(7)
	if c.String() != "7" {
		t.Errorf("String() = %v, want '7'", c.String())
	}
}

func TestContribution_Equal(t *testing.T) {
	c1 := MustParse(5)
	c2 := MustParse(5)
	c3 := MustParse(3)

	if !c1.Equal(c2) {
		t.Error("Equal() = false for equal contributions")
	}
	if c1.Equal(c3) {
		t.Error("Equal() = true for different contributions")
	}
}

func TestContribution_MarshalText(t *testing.T) {
	c := MustParse(8)
	data, err := c.MarshalText()
	if err != nil {
		t.Errorf("MarshalText() error = %v", err)
		return
	}
	if string(data) != "8" {
		t.Errorf("MarshalText() = %v, want '8'", string(data))
	}
}

func TestContribution_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:    "valid value",
			input:   "5",
			want:    5,
			wantErr: false,
		},
		{
			name:    "boundary min",
			input:   "0",
			want:    0,
			wantErr: false,
		},
		{
			name:    "boundary max",
			input:   "10",
			want:    10,
			wantErr: false,
		},
		{
			name:    "invalid not a number",
			input:   "abc",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid out of range",
			input:   "11",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid negative",
			input:   "-1",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c Contribution
			err := c.UnmarshalText([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && c.Value() != tt.want {
				t.Errorf("UnmarshalText() = %v, want %v", c.Value(), tt.want)
			}
		})
	}
}

// TestSemanticMeaning verifies the documented semantic meaning:
// 0 = tracked but doesn't contribute numerically to the objective.
func TestSemanticMeaning(t *testing.T) {
	t.Run("zero is valid and means no contribution", func(t *testing.T) {
		c, err := Parse(0)
		if err != nil {
			t.Errorf("Parse(0) should be valid: %v", err)
		}
		if c.Value() != 0 {
			t.Errorf("Parse(0).Value() = %v, want 0", c.Value())
		}
	})
}
