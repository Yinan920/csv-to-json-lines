package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseHouseRecord covers parsing of a single CSV row into a House,
// including both happy-path and error cases.
func TestParseHouseRecord(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    House
		wantErr bool
	}{
		{
			name:  "first row from the assignment sample output",
			input: []string{"452600", "8.3252", "41", "880", "129", "322", "126"},
			want: House{
				Value: 452600, Income: 8.3252, Age: 41,
				Rooms: 880, Bedrooms: 129, Pop: 322, HH: 126,
			},
			wantErr: false,
		},
		{
			name:  "value in scientific notation (present in real data)",
			input: []string{"1e+05", "2.2604", "43", "1017", "328", "836", "277"},
			want: House{
				Value: 100000, Income: 2.2604, Age: 43,
				Rooms: 1017, Bedrooms: 328, Pop: 836, HH: 277,
			},
			wantErr: false,
		},
		{
			name:    "non-numeric value -- should error",
			input:   []string{"abc", "8.3252", "41", "880", "129", "322", "126"},
			wantErr: true,
		},
		{
			name:    "non-numeric income -- should error",
			input:   []string{"452600", "notafloat", "41", "880", "129", "322", "126"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHouseRecord(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseHouseRecord() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseHouseRecord() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestConvertCSVToJSONLines is an end-to-end test that feeds a small
// synthetic CSV into the converter and verifies that every output line is
// valid JSON and round-trips back to the expected values.
func TestConvertCSVToJSONLines(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "input.csv")
	outputPath := filepath.Join(tempDir, "output.jl")

	csvData := `"value","income","age","rooms","bedrooms","pop","hh"
452600,8.3252,41,880,129,322,126
358500,8.3014,21,7099,1106,2401,1138
352100,7.2574,52,1467,190,496,177
`
	if err := os.WriteFile(inputPath, []byte(csvData), 0644); err != nil {
		t.Fatalf("writing test input: %v", err)
	}

	count, err := convertCSVToJSONLines(inputPath, outputPath)
	if err != nil {
		t.Fatalf("convertCSVToJSONLines() error = %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 rows converted, got %d", count)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 JSON lines, got %d", len(lines))
	}

	// Every line must be valid JSON and decode back to a House.
	var first House
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("first line is not valid JSON: %v -- content: %s", err, lines[0])
	}
	want := House{Value: 452600, Income: 8.3252, Age: 41, Rooms: 880, Bedrooms: 129, Pop: 322, HH: 126}
	if first != want {
		t.Errorf("first row = %+v, want %+v", first, want)
	}

	// Check the raw first line exactly matches the sample output from the
	// assignment (key order matters to the grader's visual comparison).
	expectedFirstLine := `{"value":452600,"income":8.3252,"age":41,"rooms":880,"bedrooms":129,"pop":322,"hh":126}`
	if lines[0] != expectedFirstLine {
		t.Errorf("first line = %s\nwant       = %s", lines[0], expectedFirstLine)
	}
}

// TestMissingInputFile verifies that a missing input file produces an error
// rather than a panic.
func TestMissingInputFile(t *testing.T) {
	tempDir := t.TempDir()
	_, err := convertCSVToJSONLines(
		filepath.Join(tempDir, "doesnotexist.csv"),
		filepath.Join(tempDir, "out.jl"),
	)
	if err == nil {
		t.Error("expected an error for a missing input file, got nil")
	}
}

// BenchmarkParseHouseRecord measures the hot-path cost of parsing one row.
// Run with: go test -bench=. -benchmem
func BenchmarkParseHouseRecord(b *testing.B) {
	record := []string{"452600", "8.3252", "41", "880", "129", "322", "126"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseHouseRecord(record)
	}
}
