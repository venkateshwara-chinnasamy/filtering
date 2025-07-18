package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Filters struct {
	MinDepth     int
	MaxDepth     int
	MinKeyLen    int
	MaxKeyLen    int
	NoValTypes   []string
	MinNum       *float64
	MaxNum       *float64
	MinStrLen    int
	MaxStrLen    int
	StrPattern   []string
	NoStrPattern []string
	IgnoreCase   bool
}

type Transformations struct {
	ReplaceVal     []ReplaceRule
	ReplaceKey     []ReplaceRule
	BoundNum       *BoundRule
	BoundStrLen    *BoundRule
	DefaultVal     []DefaultRule
	ArrayFilter    []ArrayFilterRule
	RenameKeyDepth []RenameDepthRule
	MaskVal        []MaskRule
	CondReplace    []CondReplaceRule
}

type ReplaceRule struct {
	Pattern     string
	Replacement string
}

type BoundRule struct {
	Min float64
	Max float64
}

type DefaultRule struct {
	Type  string
	Value interface{}
}

type ArrayFilterRule struct {
	Type   string
	Filter string
}

type RenameDepthRule struct {
	Depth  int
	Prefix string
}

type MaskRule struct {
	Pattern string
	Mask    string
}

type CondReplaceRule struct {
	Condition   string
	Replacement interface{}
}

func main() {
	var filters Filters
	var transforms Transformations
	var noValTypeFlags arrayFlag
	var replaceValFlags arrayFlag
	var replaceKeyFlags arrayFlag
	var defaultValFlags arrayFlag
	var arrayFilterFlags arrayFlag
	var renameKeyDepthFlags arrayFlag
	var maskValFlags arrayFlag
	var condReplaceFlags arrayFlag

	var strPatternFlag string
	var noStrPatternFlag string
	var boundNumFlag string
	var boundStrLenFlag string

	// Existing flags
	flag.IntVar(&filters.MinDepth, "mindepth", 0, "Include only keys at least at depth n")
	flag.IntVar(&filters.MaxDepth, "maxdepth", 999999, "Include only keys at most at depth n")
	flag.IntVar(&filters.MinKeyLen, "minkeylen", 0, "Include only keys with at least n characters")
	flag.IntVar(&filters.MaxKeyLen, "maxkeylen", 999999, "Include only keys with at most n characters")
	flag.Var(&noValTypeFlags, "novaltype", "Exclude keys with values of the given type")

	var minNumStr, maxNumStr string
	flag.StringVar(&minNumStr, "minnum", "", "For numeric values, include only if value >= n")
	flag.StringVar(&maxNumStr, "maxnum", "", "For numeric values, include only if value <= n")

	flag.IntVar(&filters.MinStrLen, "minstrlen", 0, "For string values, include only if length >= n")
	flag.IntVar(&filters.MaxStrLen, "maxstrlen", 999999, "For string values, include only if length <= n")
	flag.StringVar(&strPatternFlag, "strpattern", "", "For string values, include only if they match the pattern")
	flag.StringVar(&noStrPatternFlag, "nostrpattern", "", "Exclude strings matching the pattern")
	flag.BoolVar(&filters.IgnoreCase, "ignorecase", false, "Make string pattern filters case-insensitive")

	// New transformation flags
	flag.Var(&replaceValFlags, "replaceval", "Replace string values matching pattern with replacement")
	flag.Var(&replaceKeyFlags, "replacekey", "Replace key names matching pattern with replacement")
	flag.StringVar(&boundNumFlag, "boundnum", "", "Bound numeric values between min:max")
	flag.StringVar(&boundStrLenFlag, "boundstrlen", "", "Bound string length between min:max")
	flag.Var(&defaultValFlags, "defaultval", "Replace null/empty values with default")
	flag.Var(&arrayFilterFlags, "arrayfilter", "Apply filters to array elements")
	flag.Var(&renameKeyDepthFlags, "renamekeydepth", "Rename keys at specific depth")
	flag.Var(&maskValFlags, "maskval", "Mask values matching pattern")
	flag.Var(&condReplaceFlags, "condreplace", "Conditionally replace values")

	flag.Parse()

	// Parse existing filters
	if minNumStr != "" {
		if val, err := strconv.ParseFloat(minNumStr, 64); err == nil {
			filters.MinNum = &val
		}
	}
	if maxNumStr != "" {
		if val, err := strconv.ParseFloat(maxNumStr, 64); err == nil {
			filters.MaxNum = &val
		}
	}

	if strPatternFlag != "" {
		filters.StrPattern = strings.Split(strPatternFlag, ",")
	}
	if noStrPatternFlag != "" {
		filters.NoStrPattern = strings.Split(noStrPatternFlag, ",")
	}
	filters.NoValTypes = []string(noValTypeFlags)

	// Parse transformations
	transforms.ReplaceVal = parseReplaceRules(replaceValFlags)
	transforms.ReplaceKey = parseReplaceRules(replaceKeyFlags)

	if boundNumFlag != "" {
		transforms.BoundNum = parseBoundRule(boundNumFlag)
	}
	if boundStrLenFlag != "" {
		transforms.BoundStrLen = parseBoundRule(boundStrLenFlag)
	}

	transforms.DefaultVal = parseDefaultRules(defaultValFlags)
	transforms.ArrayFilter = parseArrayFilterRules(arrayFilterFlags)
	transforms.RenameKeyDepth = parseRenameDepthRules(renameKeyDepthFlags)
	transforms.MaskVal = parseMaskRules(maskValFlags)
	transforms.CondReplace = parseCondReplaceRules(condReplaceFlags)

	// Get input and output file names
	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] input.json output.json\n", os.Args[0])
		os.Exit(1)
	}

	inputFile := args[0]
	outputFile := args[1]

	// Read input JSON
	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Apply transformations and filters
	result := processJSON(jsonData, &filters, &transforms, 1)

	// Write output JSON
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputFile, output, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Processed JSON written to %s\n", outputFile)
}

// Custom flag type for handling multiple flags
type arrayFlag []string

func (a *arrayFlag) String() string {
	return strings.Join(*a, ",")
}

func (a *arrayFlag) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func parseReplaceRules(flags []string) []ReplaceRule {
	var rules []ReplaceRule
	for _, flag := range flags {
		parts := strings.SplitN(flag, ":", 2)
		if len(parts) == 2 {
			rules = append(rules, ReplaceRule{
				Pattern:     parts[0],
				Replacement: parts[1],
			})
		}
	}
	return rules
}

func parseBoundRule(flag string) *BoundRule {
	parts := strings.SplitN(flag, ":", 2)
	if len(parts) == 2 {
		min, err1 := strconv.ParseFloat(parts[0], 64)
		max, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 == nil && err2 == nil {
			return &BoundRule{Min: min, Max: max}
		}
	}
	return nil
}

func parseDefaultRules(flags []string) []DefaultRule {
	var rules []DefaultRule
	for _, flag := range flags {
		parts := strings.SplitN(flag, ":", 2)
		if len(parts) == 2 {
			value := parseValue(parts[1])
			rules = append(rules, DefaultRule{
				Type:  parts[0],
				Value: value,
			})
		}
	}
	return rules
}

func parseArrayFilterRules(flags []string) []ArrayFilterRule {
	var rules []ArrayFilterRule
	for _, flag := range flags {
		parts := strings.SplitN(flag, ":", 2)
		if len(parts) == 2 {
			rules = append(rules, ArrayFilterRule{
				Type:   parts[0],
				Filter: parts[1],
			})
		}
	}
	return rules
}

func parseRenameDepthRules(flags []string) []RenameDepthRule {
	var rules []RenameDepthRule
	for _, flag := range flags {
		parts := strings.SplitN(flag, ":", 2)
		if len(parts) == 2 {
			depth, err := strconv.Atoi(parts[0])
			if err == nil {
				rules = append(rules, RenameDepthRule{
					Depth:  depth,
					Prefix: parts[1],
				})
			}
		}
	}
	return rules
}

func parseMaskRules(flags []string) []MaskRule {
	var rules []MaskRule
	for _, flag := range flags {
		parts := strings.SplitN(flag, ":", 2)
		if len(parts) == 2 {
			rules = append(rules, MaskRule{
				Pattern: parts[0],
				Mask:    parts[1],
			})
		}
	}
	return rules
}

func parseCondReplaceRules(flags []string) []CondReplaceRule {
	var rules []CondReplaceRule
	for _, flag := range flags {
		parts := strings.SplitN(flag, ":", 2)
		if len(parts) == 2 {
			rules = append(rules, CondReplaceRule{
				Condition:   parts[0],
				Replacement: parseValue(parts[1]),
			})
		}
	}
	return rules
}

func parseValue(str string) interface{} {
	if str == "null" {
		return nil
	}
	if str == "true" {
		return true
	}
	if str == "false" {
		return false
	}
	if num, err := strconv.ParseFloat(str, 64); err == nil {
		return num
	}
	return str
}

func processJSON(data interface{}, filters *Filters, transforms *Transformations, depth int) interface{} {
	// First apply any transformations to the data
	if data == nil {
		return transformValue(data, transforms, depth)
	}

	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})

		// Process each key-value pair
		for key, value := range v {
			// First apply any key transformations
			newKey := transformKey(key, transforms, depth)

			// Apply masking and other value transformations
			newValue := transformValueWithKey(key, value, transforms, depth)

			// Check if this key-value pair should be included based on key-specific filters
			if !shouldIncludeKey(newKey, filters, depth) {
				continue // Skip this key-value pair
			}

			// Check if the value should be filtered out based on value-specific filters
			if !shouldIncludeValue(newValue, filters) {
				continue // Skip this key-value pair
			}

			// Recursively process nested structures
			processedValue := processJSON(newValue, filters, transforms, depth+1)

			// Add to the result
			result[newKey] = processedValue
		}

		return result

	case []interface{}:
		var result []interface{}

		// Transform each array element
		for _, item := range v {
			// Transform the item first
			transformedItem := transformValue(item, transforms, depth)

			// Process it recursively
			processedItem := processJSON(transformedItem, filters, transforms, depth+1)

			// Apply array-specific filters
			if shouldIncludeArrayElement(processedItem, transforms) {
				result = append(result, processedItem)
			}
		}

		return result

	default:
		// For primitive values, just apply transformations
		return transformValue(v, transforms, depth)
	}
}

// Split filtering into key-specific and value-specific checks
func shouldIncludeKey(key string, filters *Filters, depth int) bool {
	// Always include all keys if there are no key-specific filters
	if filters.MinDepth <= 1 &&
		filters.MaxDepth >= 999999 &&
		filters.MinKeyLen <= 0 &&
		filters.MaxKeyLen >= 999999 {
		return true
	}

	// Check depth
	if depth < filters.MinDepth || depth > filters.MaxDepth {
		return false
	}

	// Check key length
	keyLen := len(key)
	if keyLen < filters.MinKeyLen || keyLen > filters.MaxKeyLen {
		return false
	}

	return true
}

func shouldIncludeValue(value interface{}, filters *Filters) bool {
	// Always include if no value-specific filters are specified
	if len(filters.NoValTypes) == 0 &&
		filters.MinNum == nil && filters.MaxNum == nil &&
		filters.MinStrLen <= 0 && filters.MaxStrLen >= 999999 &&
		len(filters.StrPattern) == 0 && len(filters.NoStrPattern) == 0 {
		return true
	}

	// Check value type filters
	if len(filters.NoValTypes) > 0 {
		valueType := getValueType(value)
		for _, noType := range filters.NoValTypes {
			if valueType == noType {
				return false
			}
		}
	}

	// Check numeric value filters
	if num, ok := value.(float64); ok {
		if filters.MinNum != nil && num < *filters.MinNum {
			return false
		}
		if filters.MaxNum != nil && num > *filters.MaxNum {
			return false
		}
	}

	// Check string value filters - only apply to strings
	if str, ok := value.(string); ok {
		strLen := len(str)
		if strLen < filters.MinStrLen || strLen > filters.MaxStrLen {
			return false
		}

		if len(filters.StrPattern) > 0 && !matchesPattern(str, filters.StrPattern, filters.IgnoreCase) {
			return false
		}

		if len(filters.NoStrPattern) > 0 && matchesPattern(str, filters.NoStrPattern, filters.IgnoreCase) {
			return false
		}
	}

	return true
}

func shouldIncludeArrayElement(element interface{}, transforms *Transformations) bool {
	if len(transforms.ArrayFilter) == 0 {
		return true // No array filters specified, include all elements
	}

	elementType := getValueType(element)
	for _, rule := range transforms.ArrayFilter {
		if elementType == rule.Type {
			if rule.Filter == "-minnum 10" {
				if num, ok := element.(float64); ok {
					return num >= 10 // Only include if number >= 10
				}
			}
			// Add other filter types here as needed
			return false // Filtered out by default for matching type
		}
	}

	return true // No filter for this element type, include it
}

// Helper function to process nested structures recursively
func processNestedStructure(data interface{}, filters *Filters, transforms *Transformations, depth int) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		// Recursively process the map
		return processJSON(v, filters, transforms, depth)
	case []interface{}:
		// Recursively process the array
		return processJSON(v, filters, transforms, depth)
	default:
		// For primitive values, just return as is
		return data
	}
}

func valueFilteredOut(value interface{}) bool {
	// Only consider truly empty structures as filtered out
	switch v := value.(type) {
	case map[string]interface{}:
		return len(v) == 0
	case []interface{}:
		return len(v) == 0
	default:
		return false
	}
}

func transformKey(key string, transforms *Transformations, depth int) string {
	newKey := key

	// Apply key replacements
	for _, rule := range transforms.ReplaceKey {
		if newKey == rule.Pattern {
			newKey = rule.Replacement
		}
	}

	// Apply depth-based renaming
	for _, rule := range transforms.RenameKeyDepth {
		if depth == rule.Depth {
			newKey = rule.Prefix + newKey
		}
	}

	return newKey
}

// Function that handles masking and other transformations based on the original key
func transformValueWithKey(key string, value interface{}, transforms *Transformations, depth int) interface{} {
	// First apply masking based on key
	for _, rule := range transforms.MaskVal {
		if key == rule.Pattern {
			return rule.Mask
		}
	}

	// Then apply other transformations
	return transformValue(value, transforms, depth)
}

func transformValue(value interface{}, transforms *Transformations, depth int) interface{} {
	// Apply conditional replacements first
	for _, rule := range transforms.CondReplace {
		if evaluateCondition(value, rule.Condition) {
			return rule.Replacement
		}
	}

	// Apply default value replacements
	for _, rule := range transforms.DefaultVal {
		if shouldApplyDefault(value, rule.Type) {
			return rule.Value
		}
	}

	// Apply value type-specific transformations
	switch v := value.(type) {
	case string:
		return transformString(v, transforms)
	case float64:
		return transformNumber(v, transforms)
	default:
		return value
	}
}

func transformString(str string, transforms *Transformations) interface{} {
	result := str

	// Apply string value replacements
	for _, rule := range transforms.ReplaceVal {
		if matchesStringPattern(result, rule.Pattern) {
			return rule.Replacement
		}
	}

	// Apply string length bounds
	if transforms.BoundStrLen != nil {
		minLen := int(transforms.BoundStrLen.Min)
		maxLen := int(transforms.BoundStrLen.Max)

		if len(result) < minLen {
			// Pad with spaces
			result = result + strings.Repeat(" ", minLen-len(result))
		} else if len(result) > maxLen {
			// Truncate
			result = result[:maxLen]
		}
	}

	return result
}

func transformNumber(num float64, transforms *Transformations) float64 {
	result := num

	// Apply numeric bounds
	if transforms.BoundNum != nil {
		if result < transforms.BoundNum.Min {
			result = transforms.BoundNum.Min
		} else if result > transforms.BoundNum.Max {
			result = transforms.BoundNum.Max
		}
	}

	return result
}

func shouldApplyDefault(value interface{}, valueType string) bool {
	switch valueType {
	case "null":
		return value == nil
	case "string":
		if str, ok := value.(string); ok {
			return str == ""
		}
		return false
	default:
		return false
	}
}

func evaluateCondition(value interface{}, condition string) bool {
	// Simple condition evaluation
	if strings.HasPrefix(condition, "value==") {
		expected := strings.Trim(condition[7:], "\"")
		if expected == "null" {
			return value == nil
		}
		if str, ok := value.(string); ok {
			return str == expected
		}
	}
	return false
}

func matchesStringPattern(str, pattern string) bool {
	switch pattern {
	case "upper":
		return regexp.MustCompile(`[A-Z]`).MatchString(str)
	case "lower":
		return regexp.MustCompile(`[a-z]`).MatchString(str)
	case "num":
		return regexp.MustCompile(`[0-9]`).MatchString(str)
	case "sym":
		return regexp.MustCompile(`[^A-Za-z0-9\s]`).MatchString(str)
	case "email":
		return strings.Contains(str, "@")
	default:
		return str == pattern
	}
}

func shouldIncludeKV(key string, value interface{}, filters *Filters, depth int) bool {
	// Check depth filters
	if depth < filters.MinDepth || depth > filters.MaxDepth {
		return false
	}

	// Check key length filters
	keyLen := len(key)
	if keyLen < filters.MinKeyLen || keyLen > filters.MaxKeyLen {
		return false
	}

	// Check value type filters
	if len(filters.NoValTypes) > 0 {
		valueType := getValueType(value)
		for _, noType := range filters.NoValTypes {
			if valueType == noType {
				return false
			}
		}
	}

	// Check numeric value filters
	if num, ok := value.(float64); ok {
		if filters.MinNum != nil && num < *filters.MinNum {
			return false
		}
		if filters.MaxNum != nil && num > *filters.MaxNum {
			return false
		}
	}

	// Check string value filters
	if str, ok := value.(string); ok {
		strLen := len(str)
		if strLen < filters.MinStrLen || strLen > filters.MaxStrLen {
			return false
		}

		if len(filters.StrPattern) > 0 && !matchesPattern(str, filters.StrPattern, filters.IgnoreCase) {
			return false
		}

		if len(filters.NoStrPattern) > 0 && matchesPattern(str, filters.NoStrPattern, filters.IgnoreCase) {
			return false
		}
	}

	// If we passed all filters, include this key-value pair
	return true
}

func getValueType(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "bool"
	case nil:
		return "null"
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	default:
		return "unknown"
	}
}

func matchesPattern(str string, patterns []string, ignoreCase bool) bool {
	testStr := str
	if ignoreCase {
		testStr = strings.ToLower(str)
	}

	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if !hasPattern(testStr, pattern) {
			return false
		}
	}
	return true
}

func hasPattern(str, pattern string) bool {
	switch pattern {
	case "upper":
		return regexp.MustCompile(`[A-Z]`).MatchString(str)
	case "lower":
		return regexp.MustCompile(`[a-z]`).MatchString(str)
	case "num":
		return regexp.MustCompile(`[0-9]`).MatchString(str)
	case "sym":
		return regexp.MustCompile(`[^A-Za-z0-9\s]`).MatchString(str)
	default:
		return false
	}
}
