// Package main implements a command-line utility that converts a California
// housing CSV file into a JSON Lines (.jl) file.
//
// Usage:
//
//	./csv-to-json-lines <input.csv> <output.jl>
//
// Example:
//
//	./csv-to-json-lines housesInput.csv housesOutput.jl
package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

// House represents one row of California housing data from the input CSV.
// The `json:"..."` struct tags control the JSON output keys so that they
// match the field names shown in the assignment's sample output.
type House struct {
	Value    int     `json:"value"`    // median house value
	Income   float64 `json:"income"`   // median household income
	Age      int     `json:"age"`      // housing median age
	Rooms    int     `json:"rooms"`    // total rooms
	Bedrooms int     `json:"bedrooms"` // total bedrooms
	Pop      int     `json:"pop"`      // block-group population
	HH       int     `json:"hh"`       // households
}

// expectedColumns is the number of columns each row of the input CSV must have.
const expectedColumns = 7

func main() {
	// Expect exactly two arguments: the input and output paths.
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr,
			"Usage: %s <input.csv> <output.jl>\n"+
				"Example: %s housesInput.csv housesOutput.jl\n",
			os.Args[0], os.Args[0])
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]

	count, err := convertCSVToJSONLines(inputPath, outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Success: converted %d rows from %s to %s\n",
		count, inputPath, outputPath)
}

// convertCSVToJSONLines reads the input CSV file, converts each data row into
// a JSON object, and writes one JSON object per line to the output file.
// It returns the number of data rows written (excluding the header) and any
// error encountered.
func convertCSVToJSONLines(inputPath, outputPath string) (int, error) {
	inFile, err := os.Open(inputPath)
	if err != nil {
		return 0, fmt.Errorf("opening input file: %w", err)
	}
	defer inFile.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return 0, fmt.Errorf("creating output file: %w", err)
	}
	defer outFile.Close()

	reader := csv.NewReader(inFile)
	// Enforce that every row has the same number of fields as the header.
	reader.FieldsPerRecord = expectedColumns

	// Read and discard the header row.
	if _, err := reader.Read(); err != nil {
		return 0, fmt.Errorf("reading header: %w", err)
	}

	// json.Encoder writes a newline after every encoded value, which is
	// exactly what the JSON Lines format requires.
	encoder := json.NewEncoder(outFile)

	recordNum := 0
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return recordNum, fmt.Errorf("reading row %d: %w", recordNum+1, err)
		}
		recordNum++

		house, err := parseHouseRecord(record)
		if err != nil {
			return recordNum, fmt.Errorf("parsing row %d: %w", recordNum, err)
		}

		if err := encoder.Encode(house); err != nil {
			return recordNum, fmt.Errorf("writing row %d: %w", recordNum, err)
		}
	}

	return recordNum, nil
}

// parseHouseRecord converts a single slice of CSV string fields into a House
// struct, returning an error if any field cannot be parsed as the expected
// numeric type.
//
// Note: integer fields are parsed with ParseFloat (then cast to int) rather
// than Atoi, because the source data contains some values written in
// scientific notation (for example, "1e+05" means 100000). This lets us
// accept both "452600" and "1e+05" without special-casing.
func parseHouseRecord(record []string) (House, error) {
	var h House

	intField := func(field, name string) (int, error) {
		f, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return 0, fmt.Errorf("%s field %q is not a valid number: %w", name, field, err)
		}
		return int(f), nil
	}

	var err error
	if h.Value, err = intField(record[0], "value"); err != nil {
		return h, err
	}
	if h.Income, err = strconv.ParseFloat(record[1], 64); err != nil {
		return h, fmt.Errorf("income field %q is not a valid float: %w", record[1], err)
	}
	if h.Age, err = intField(record[2], "age"); err != nil {
		return h, err
	}
	if h.Rooms, err = intField(record[3], "rooms"); err != nil {
		return h, err
	}
	if h.Bedrooms, err = intField(record[4], "bedrooms"); err != nil {
		return h, err
	}
	if h.Pop, err = intField(record[5], "pop"); err != nil {
		return h, err
	}
	if h.HH, err = intField(record[6], "hh"); err != nil {
		return h, err
	}

	return h, nil
}
