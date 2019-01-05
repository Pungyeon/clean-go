package utils

import "testing"

func TestToReadableSize(t *testing.T) {
	tt := []struct {
		name     string
		input    int64
		expected string
	}{
		{"byte return", 125, "125 B"},
		{"kilobyte return", 1010, "1.01 KB"},
		{"megabyte return", 1988909, "1.99 MB"},
		{"gigabyte return", 29121988909, "29.12 GB"},
		{"gigabyte return", 890929121988909, "890.93 TB"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			output := ToReadableSize(tc.input)
			if output != tc.expected {
				t.Errorf("input %d, unexpected output: %s", tc.input, output)
			}
		})
	}
}
