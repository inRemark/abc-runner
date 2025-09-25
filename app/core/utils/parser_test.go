package utils

import (
	"testing"
)

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"123", 123, false},
		{"-456", -456, false},
		{"0", 0, false},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, test := range tests {
		result, err := ParseInt(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("ParseInt(%s) expected error but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseInt(%s) unexpected error: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("ParseInt(%s) = %d, expected %d", test.input, result, test.expected)
			}
		}
	}
}

func TestParseInt64(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"123", 123, false},
		{"-456", -456, false},
		{"0", 0, false},
		{"9223372036854775807", 9223372036854775807, false}, // Max int64
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, test := range tests {
		result, err := ParseInt64(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("ParseInt64(%s) expected error but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseInt64(%s) unexpected error: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("ParseInt64(%s) = %d, expected %d", test.input, result, test.expected)
			}
		}
	}
}

func TestParseFloat64(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"123.45", 123.45, false},
		{"-678.90", -678.90, false},
		{"0", 0.0, false},
		{"3.14159", 3.14159, false},
		{"abc", 0.0, true},
		{"", 0.0, true},
	}

	for _, test := range tests {
		result, err := ParseFloat64(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("ParseFloat64(%s) expected error but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseFloat64(%s) unexpected error: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("ParseFloat64(%s) = %f, expected %f", test.input, result, test.expected)
			}
		}
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		hasError bool
	}{
		{"true", true, false},
		{"false", false, false},
		{"1", true, false},
		{"0", false, false},
		{"t", true, false},
		{"f", false, false},
		{"abc", false, true},
		{"", false, true},
	}

	for _, test := range tests {
		result, err := ParseBool(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("ParseBool(%s) expected error but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseBool(%s) unexpected error: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("ParseBool(%s) = %t, expected %t", test.input, result, test.expected)
			}
		}
	}
}
