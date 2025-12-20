package frequencycount_test

import (
	"testing"

	"github.com/francowini/rafiki/business/types/frequencycount"
)

func TestParse_Valid(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"minimum valid value", 1, 1},
		{"small positive value", 5, 5},
		{"larger positive value", 100, 100},
		{"typical weekly frequency", 7, 7},
		{"typical monthly frequency", 30, 30},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fc, err := frequencycount.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%d) returned unexpected error: %v", tc.input, err)
			}
			if fc.Value() != tc.expected {
				t.Errorf("Parse(%d).Value() = %d, want %d", tc.input, fc.Value(), tc.expected)
			}
		})
	}
}

func TestParse_Invalid(t *testing.T) {
	tests := []struct {
		name          string
		input         int
		expectedError string
	}{
		{"zero value", 0, "frequency count must be at least 1, got 0"},
		{"negative value", -1, "frequency count must be at least 1, got -1"},
		{"large negative value", -100, "frequency count must be at least 1, got -100"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := frequencycount.Parse(tc.input)
			if err == nil {
				t.Fatalf("Parse(%d) expected error, got nil", tc.input)
			}
			if err.Error() != tc.expectedError {
				t.Errorf("Parse(%d) error = %q, want %q", tc.input, err.Error(), tc.expectedError)
			}
		})
	}
}

func TestMustParse_Valid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MustParse(5) panicked unexpectedly: %v", r)
		}
	}()

	fc := frequencycount.MustParse(5)
	if fc.Value() != 5 {
		t.Errorf("MustParse(5).Value() = %d, want 5", fc.Value())
	}
}

func TestMustParse_Invalid(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustParse(0) expected panic, got nil")
		}
	}()

	frequencycount.MustParse(0)
}

func TestValue(t *testing.T) {
	tests := []int{1, 5, 10, 100}

	for _, expected := range tests {
		fc := frequencycount.MustParse(expected)
		if fc.Value() != expected {
			t.Errorf("Value() = %d, want %d", fc.Value(), expected)
		}
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{1, "1"},
		{5, "5"},
		{100, "100"},
	}

	for _, tc := range tests {
		fc := frequencycount.MustParse(tc.input)
		if fc.String() != tc.expected {
			t.Errorf("String() = %q, want %q", fc.String(), tc.expected)
		}
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected bool
	}{
		{"equal values", 5, 5, true},
		{"different values", 5, 10, false},
		{"both minimum", 1, 1, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcA := frequencycount.MustParse(tc.a)
			fcB := frequencycount.MustParse(tc.b)
			if fcA.Equal(fcB) != tc.expected {
				t.Errorf("Equal() = %v, want %v", fcA.Equal(fcB), tc.expected)
			}
		})
	}
}

func TestMarshalText(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{1, "1"},
		{5, "5"},
		{100, "100"},
	}

	for _, tc := range tests {
		fc := frequencycount.MustParse(tc.input)
		data, err := fc.MarshalText()
		if err != nil {
			t.Fatalf("MarshalText() returned unexpected error: %v", err)
		}
		if string(data) != tc.expected {
			t.Errorf("MarshalText() = %q, want %q", string(data), tc.expected)
		}
	}
}
