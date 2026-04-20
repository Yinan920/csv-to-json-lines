# csv-to-json-lines

A Go command-line tool that converts California housing CSV data into
[JSON Lines](https://jsonlines.org/) (`.jl`) format. Submitted for Week 3:
*Creating a Command-Line Application*.

## Problem

A data science consulting firm needs a utility for converting CSV data into
JSON Lines, a common format for populating databases and document stores.
The input data comes from Miller (2015)'s study of California housing
prices.

Each output line is a single valid JSON object, separated by `\n`:

```
{"value":452600,"income":8.3252,"age":41,"rooms":880,"bedrooms":129,"pop":322,"hh":126}
{"value":358500,"income":8.3014,"age":21,"rooms":7099,"bedrooms":1106,"pop":2401,"hh":1138}
```

## Repository layout

```
csv-to-json-lines/
├── go.mod              # Go module definition
├── main.go             # CLI entry point and conversion logic
├── main_test.go        # Unit tests and a benchmark
├── README.md           # This file
├── .gitignore
└── testdata/
    └── sample.csv      # 5-row sample for a quick smoke test
```

## Requirements

- [Go](https://go.dev/dl/) 1.22 or newer

Verify with:

```sh
go version
```

## Build

From the repository root:

```sh
# macOS / Linux
go build -o csv-to-json-lines .

# Windows (PowerShell)
go build -o csv-to-json-lines.exe .
```

This produces a standalone executable (`csv-to-json-lines` on macOS/Linux,
`csv-to-json-lines.exe` on Windows) that runs without Go installed.

## Usage

The program takes two command-line arguments: the input CSV path and the
output JSON Lines path.

```sh
./csv-to-json-lines <input.csv> <output.jl>
```

Example:

```sh
# macOS / Linux
./csv-to-json-lines housesInput.csv housesOutput.jl

# Windows
.\csv-to-json-lines.exe housesInput.csv housesOutput.jl
```

On success the program prints the number of rows it converted:

```
Success: converted 20640 rows from housesInput.csv to housesOutput.jl
```

### Error handling

- Wrong number of arguments → prints a usage message and exits with status 1.
- Input file missing or unreadable → prints an error with the offending path.
- CSV row has the wrong number of fields → reports the row number.
- A field cannot be parsed as a number → reports the row number and the
  field name.

### Input schema

The input CSV must have exactly 7 columns with this header:

```csv
"value","income","age","rooms","bedrooms","pop","hh"
```

| Column     | Type    | Description                         |
|------------|---------|-------------------------------------|
| `value`    | integer | Median house value (US dollars)     |
| `income`   | float   | Median household income             |
| `age`      | integer | Housing median age (years)          |
| `rooms`    | integer | Total rooms                         |
| `bedrooms` | integer | Total bedrooms                      |
| `pop`      | integer | Block-group population              |
| `hh`       | integer | Households                          |

**On scientific notation**: the real-world data contains 192 rows where
`value` is written in scientific notation (for example, `1e+05`). The
program parses every integer field with `ParseFloat` and then casts to
`int`, so both `452600` and `1e+05` are accepted.

## Testing

### Run the unit tests

```sh
go test -v ./...
```

Expected output (all tests pass):

```
=== RUN   TestParseHouseRecord
--- PASS: TestParseHouseRecord (0.00s)
=== RUN   TestConvertCSVToJSONLines
--- PASS: TestConvertCSVToJSONLines (0.00s)
=== RUN   TestMissingInputFile
--- PASS: TestMissingInputFile (0.00s)
PASS
```

### Run the benchmark

```sh
go test -bench=. -benchmem ./...
```

Typical result (roughly 100 ns per row, zero allocations):

```
BenchmarkParseHouseRecord-2   10865638   103.2 ns/op   0 B/op   0 allocs/op
```

### Smoke test with the small sample

```sh
./csv-to-json-lines testdata/sample.csv /tmp/out.jl
cat /tmp/out.jl
```

### Full dataset

Place the instructor-provided `housesInput.csv` in the repository root and
then:

```sh
./csv-to-json-lines housesInput.csv housesOutput.jl
wc -l housesOutput.jl        # 20640
head -3 housesOutput.jl       # matches the assignment's sample output
```

## Code quality

The project uses the standard Go toolchain:

```sh
go fmt ./...    # format
go vet ./...    # static analysis
go test ./...   # tests
```

Design notes:

- `main` has no business logic: it parses arguments, calls
  `convertCSVToJSONLines`, and reports the result.
- **Typed output**: the `House` struct declares the expected numeric type
  for every field, rather than treating everything as a string.
- **Tag-driven JSON**: `json:"..."` struct tags pin the output keys and
  casing, which removes the need to hand-build JSON strings and makes the
  output deterministic.
- **Streaming I/O**: rows are read with `encoding/csv` one at a time and
  written with `json.Encoder`, so memory use does not grow with input size.
  The 20,640-row reference file converts in milliseconds.
- **Wrapped errors**: every error is wrapped with `fmt.Errorf("... %w",
  err)` so that the final message includes useful context such as the row
  number and the field name.

## (Optional) Making it general-purpose

Supporting arbitrary CSV schemas means giving up the typed `House` struct
in favor of `map[string]any`:

```go
headers, _ := reader.Read()
for {
    record, err := reader.Read()
    if err == io.EOF { break }
    obj := make(map[string]any, len(headers))
    for i, h := range headers {
        obj[h] = record[i]   // or try to ParseFloat / ParseInt first
    }
    encoder.Encode(obj)
}
```

The trade-off is that every field becomes a string (or requires attempting
to parse as a number), losing the compile-time guarantees of the current
version. This submission targets the California housing schema
specifically, which is the right choice for this use case.

## References

- Bodner, Jon. 2024. *Learning Go: An Idiomatic Approach to Real-World Go
  Programming*, 2nd ed.
- Gerardi, Ricardo. 2021. *Powerful Command-Line Applications in Go*.
- Miller, Thomas W. 2015. *Modeling Techniques in Predictive Analytics
  with Python and R*, Chapter 10.
- Go standard library: [`encoding/csv`](https://pkg.go.dev/encoding/csv),
  [`encoding/json`](https://pkg.go.dev/encoding/json).

## Use of AI assistants

*(Replace this section with an honest description of how you actually used
AI help — the template below is a starting point.)*

Anthropic's Claude was used as a pair-programming assistant for this
assignment. Specifically, Claude helped to:

- Discuss the overall design, particularly the use of `json.Encoder` to
  produce JSON Lines output (each encoded object is automatically followed
  by a newline).
- Draft unit tests covering the happy path, malformed input, and the
  scientific-notation (`1e+05`) case that appears in the real data.
- Draft this README and code comments.

I reviewed every line of code, ran `go fmt`, `go vet`, and `go test`
locally, and verified the final program against `housesInput.csv` with
[JSONLint](https://jsonlint.com/) to confirm each line is valid JSON.
