package mathutils

import (
	"testing"
)

func TestSafeDivide(t *testing.T) {
	tests := []struct {
		name    string
		a       float64
		b       float64
		wantOk  bool
		wantVal float64
	}{
		{"valid division", 10.0, 2.0, true, 5.0},
		{"division by zero", 10.0, 0.0, false, 0.0},
		{"negative result", -10.0, 2.0, true, -5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeDivide(tt.a, tt.b)

			if result.IsOk() != tt.wantOk {
				t.Errorf("SafeDivide(%v, %v).IsOk() = %v, want %v",
					tt.a, tt.b, result.IsOk(), tt.wantOk)
			}

			if tt.wantOk && result.Unwrap() != tt.wantVal {
				t.Errorf("SafeDivide(%v, %v).Unwrap() = %v, want %v",
					tt.a, tt.b, result.Unwrap(), tt.wantVal)
			}
		})
	}
}

func TestSafeSqrt(t *testing.T) {
	tests := []struct {
		name    string
		x       float64
		wantOk  bool
		wantVal float64
	}{
		{"positive number", 16.0, true, 4.0},
		{"zero", 0.0, true, 0.0},
		{"negative number", -16.0, false, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeSqrt(tt.x)

			if result.IsOk() != tt.wantOk {
				t.Errorf("SafeSqrt(%v).IsOk() = %v, want %v",
					tt.x, result.IsOk(), tt.wantOk)
			}

			if tt.wantOk && result.Unwrap() != tt.wantVal {
				t.Errorf("SafeSqrt(%v).Unwrap() = %v, want %v",
					tt.x, result.Unwrap(), tt.wantVal)
			}
		})
	}
}

func TestSafeModulo(t *testing.T) {
	tests := []struct {
		name    string
		a       int
		b       int
		wantOk  bool
		wantVal int
	}{
		{"valid modulo", 10, 3, true, 1},
		{"modulo by zero", 10, 0, false, 0},
		{"exact division", 10, 5, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeModulo(tt.a, tt.b)

			if result.IsOk() != tt.wantOk {
				t.Errorf("SafeModulo(%v, %v).IsOk() = %v, want %v",
					tt.a, tt.b, result.IsOk(), tt.wantOk)
			}

			if tt.wantOk && result.Unwrap() != tt.wantVal {
				t.Errorf("SafeModulo(%v, %v).Unwrap() = %v, want %v",
					tt.a, tt.b, result.Unwrap(), tt.wantVal)
			}
		})
	}
}
