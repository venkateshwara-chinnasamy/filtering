# filtering

# Run all tests
go test -v

# Run specific test
go test -v -run TestReplaceVal

# Run benchmarks
go test -bench=.

# Run with coverage
go test -cover


The program now supports all the requested transformation features:
-replaceval: Replaces string values matching patterns
-replacekey: Replaces key names
-boundnum: Bounds numeric values between min and max
-boundstrlen: Bounds string length with padding/truncation
-defaultval: Replaces null/empty values with defaults
-arrayfilter: Filters array elements based on type and criteria
-renamekeydepth: Renames keys at specific depths
-maskval: Masks values based on key patterns
-condreplace: Conditionally replaces values
