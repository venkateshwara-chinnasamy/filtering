package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// Test helper functions
func createTestInput() map[string]interface{} {
	return map[string]interface{}{
		"Name":  "Alice",
		"age":   30.0,
		"email": "ALICE@EXAMPLE.COM",
		"score": 99.5,
		"meta": map[string]interface{}{
			"verified": true,
			"tags":     []interface{}{"VIP", "2024"},
			"profile": map[string]interface{}{
				"bio":   "Senior DEV!",
				"id":    12345.0,
				"notes": nil,
			},
		},
		"notes":    nil,
		"emptyStr": "",
		"zero":     0.0,
		"arr":      []interface{}{1.0, 2.0, 3.0},
		"SYM":      "#@!$",
		"lower":    "lowercase",
		"MIX":      "AbC123!@#",
	}
}

func writeJSONFile(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

func readJSONFile(filename string) (interface{}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func TestReplaceVal(t *testing.T) {
	input := createTestInput()

	transforms := &Transformations{
		ReplaceVal: []ReplaceRule{
			{Pattern: "upper", Replacement: "MASKED"},
		},
	}
	filters := &Filters{MaxDepth: 999999, MaxKeyLen: 999999, MaxStrLen: 999999}

	result := processJSON(input, filters, transforms, 1)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Check that uppercase strings were replaced
	if resultMap["Name"] != "MASKED" {
		t.Errorf("Expected Name to be MASKED, got %v", resultMap["Name"])
	}

	if resultMap["email"] != "MASKED" {
		t.Errorf("Expected email to be MASKED, got %v", resultMap["email"])
	}

	// Check that non-uppercase strings were not replaced
	if resultMap["lower"] != "lowercase" {
		t.Errorf("Expected lower to remain unchanged, got %v", resultMap["lower"])
	}

	// Check that MIX was replaced (contains uppercase)
	if resultMap["MIX"] != "MASKED" {
		t.Errorf("Expected MIX to be MASKED, got %v", resultMap["MIX"])
	}
}

func TestBoundNum(t *testing.T) {
	input := createTestInput()

	transforms := &Transformations{
		BoundNum: &BoundRule{Min: 10, Max: 100},
	}
	filters := &Filters{MaxDepth: 999999, MaxKeyLen: 999999, MaxStrLen: 999999}

	result := processJSON(input, filters, transforms, 1)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Check that zero was bounded to 10
	if resultMap["zero"] != 10.0 {
		t.Errorf("Expected zero to be bounded to 10, got %v", resultMap["zero"])
	}

	// Check that values within bounds remain unchanged
	if resultMap["age"] != 30.0 {
		t.Errorf("Expected age to remain 30, got %v", resultMap["age"])
	}

	// Check nested values
	meta := resultMap["meta"].(map[string]interface{})
	profile := meta["profile"].(map[string]interface{})
	if profile["id"] != 100.0 { // 12345 should be bounded to 100
		t.Errorf("Expected id to be bounded to 100, got %v", profile["id"])
	}

	// Check array elements
	arr := resultMap["arr"].([]interface{})
	for i, val := range arr {
		if val != 10.0 { // All values [1,2,3] should be bounded to 10
			t.Errorf("Expected arr[%d] to be bounded to 10, got %v", i, val)
		}
	}
}

func TestReplaceKey(t *testing.T) {
	input := createTestInput()

	transforms := &Transformations{
		ReplaceKey: []ReplaceRule{
			{Pattern: "email", Replacement: "contact"},
			{Pattern: "score", Replacement: "points"},
		},
	}
	filters := &Filters{MaxDepth: 999999, MaxKeyLen: 999999, MaxStrLen: 999999}

	result := processJSON(input, filters, transforms, 1)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Check that keys were replaced
	if _, exists := resultMap["contact"]; !exists {
		t.Error("Expected contact key to exist")
	}

	if _, exists := resultMap["points"]; !exists {
		t.Error("Expected points key to exist")
	}

	// Check that old keys don't exist
	if _, exists := resultMap["email"]; exists {
		t.Error("Expected email key to be replaced")
	}

	if _, exists := resultMap["score"]; exists {
		t.Error("Expected score key to be replaced")
	}
}

func TestBoundStrLen(t *testing.T) {
	input := createTestInput()

	transforms := &Transformations{
		BoundStrLen: &BoundRule{Min: 5, Max: 8},
	}
	filters := &Filters{MaxDepth: 999999, MaxKeyLen: 999999, MaxStrLen: 999999}

	result := processJSON(input, filters, transforms, 1)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Check that short strings were padded
	emptyStr := resultMap["emptyStr"].(string)
	if len(emptyStr) != 5 {
		t.Errorf("Expected emptyStr to be padded to length 5, got length %d", len(emptyStr))
	}

	// Check that long strings were truncated
	email := resultMap["email"].(string)
	if len(email) != 8 {
		t.Errorf("Expected email to be truncated to length 8, got length %d", len(email))
	}

	// Check that strings within bounds remain unchanged in length
	name := resultMap["Name"].(string)
	if len(name) != 5 { // "Alice" should remain as is since it's exactly 5 chars
		t.Errorf("Expected Name to remain length 5, got length %d", len(name))
	}
}

func TestDefaultVal(t *testing.T) {
	input := createTestInput()

	transforms := &Transformations{
		DefaultVal: []DefaultRule{
			{Type: "null", Value: "N/A"},
			{Type: "string", Value: "EMPTY"},
		},
	}
	filters := &Filters{MaxDepth: 999999, MaxKeyLen: 999999, MaxStrLen: 999999}

	result := processJSON(input, filters, transforms, 1)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Check that null values were replaced
	if resultMap["notes"] != "N/A" {
		t.Errorf("Expected notes to be N/A, got %v", resultMap["notes"])
	}

	// Check that empty strings were replaced
	if resultMap["emptyStr"] != "EMPTY" {
		t.Errorf("Expected emptyStr to be EMPTY, got %v", resultMap["emptyStr"])
	}

	// Check nested null values
	meta := resultMap["meta"].(map[string]interface{})
	profile := meta["profile"].(map[string]interface{})
	if profile["notes"] != "N/A" {
		t.Errorf("Expected nested notes to be N/A, got %v", profile["notes"])
	}
}

func TestMaskVal(t *testing.T) {
	input := createTestInput()

	transforms := &Transformations{
		MaskVal: []MaskRule{
			{Pattern: "email", Mask: "***MASKED***"},
			{Pattern: "Name", Mask: "MASK"},
		},
	}
	filters := &Filters{MaxDepth: 999999, MaxKeyLen: 999999, MaxStrLen: 999999}

	result := processJSON(input, filters, transforms, 1)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Check that values were masked
	if resultMap["email"] != "***MASKED***" {
		t.Errorf("Expected email to be masked with ***MASKED***, got %v", resultMap["email"])
	}

	if resultMap["Name"] != "MASK" {
		t.Errorf("Expected Name to be masked with MASK, got %v", resultMap["Name"])
	}

	// Check that other values weren't masked
	if resultMap["age"] != 30.0 {
		t.Errorf("Expected age to remain unchanged, got %v", resultMap["age"])
	}
}

func TestCondReplace(t *testing.T) {
	input := createTestInput()

	transforms := &Transformations{
		CondReplace: []CondReplaceRule{
			{Condition: "value==\"Alice\"", Replacement: "User"},
			{Condition: "value==null", Replacement: "Unknown"},
		},
	}
	filters := &Filters{MaxDepth: 999999, MaxKeyLen: 999999, MaxStrLen: 999999}

	result := processJSON(input, filters, transforms, 1)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Check that conditional replacements worked
	if resultMap["Name"] != "User" {
		t.Errorf("Expected Name to be User, got %v", resultMap["Name"])
	}

	if resultMap["notes"] != "Unknown" {
		t.Errorf("Expected notes to be Unknown, got %v", resultMap["notes"])
	}

	// Check nested null values
	meta := resultMap["meta"].(map[string]interface{})
	profile := meta["profile"].(map[string]interface{})
	if profile["notes"] != "Unknown" {
		t.Errorf("Expected nested notes to be Unknown, got %v", profile["notes"])
	}
}

func TestRenameKeyDepth(t *testing.T) {
	input := createTestInput()

	transforms := &Transformations{
		RenameKeyDepth: []RenameDepthRule{
			{Depth: 2, Prefix: "sub_"},
		},
	}
	filters := &Filters{MaxDepth: 999999, MaxKeyLen: 999999, MaxStrLen: 999999}

	result := processJSON(input, filters, transforms, 1)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Check depth 2 keys were renamed
	meta := resultMap["meta"].(map[string]interface{})
	if _, exists := meta["sub_verified"]; !exists {
		t.Error("Expected sub_verified key to exist at depth 2")
	}

	if _, exists := meta["sub_tags"]; !exists {
		t.Error("Expected sub_tags key to exist at depth 2")
	}

	if _, exists := meta["sub_profile"]; !exists {
		t.Error("Expected sub_profile key to exist at depth 2")
	}

	// Check that depth 1 keys were not renamed
	if _, exists := resultMap["meta"]; !exists {
		t.Error("Expected meta key to remain unchanged at depth 1")
	}
}

func TestArrayFilter(t *testing.T) {
	input := createTestInput()

	transforms := &Transformations{
		ArrayFilter: []ArrayFilterRule{
			{Type: "number", Filter: "-minnum 10"},
		},
	}
	filters := &Filters{MaxDepth: 999999, MaxKeyLen: 999999, MaxStrLen: 999999}

	result := processJSON(input, filters, transforms, 1)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Check that the array key exists (this was the issue)
	arr, exists := resultMap["arr"]
	if !exists {
		t.Fatal("Expected arr key to exist")
	}

	arrSlice, ok := arr.([]interface{})
	if !ok {
		t.Fatalf("Expected arr to be a slice, got %T", arr)
	}

	if len(arrSlice) != 0 { // All elements [1,2,3] should be filtered out as they're < 10
		t.Errorf("Expected empty array, got %v", arrSlice)
	}
}

func TestCombinedTransformations(t *testing.T) {
	input := createTestInput()

	transforms := &Transformations{
		ReplaceVal: []ReplaceRule{
			{Pattern: "num", Replacement: "REDACTED"},
		},
		BoundNum: &BoundRule{Min: 0, Max: 100},
		DefaultVal: []DefaultRule{
			{Type: "null", Value: 0.0},
		},
	}
	filters := &Filters{MaxDepth: 999999, MaxKeyLen: 999999, MaxStrLen: 999999}

	result := processJSON(input, filters, transforms, 1)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Check that null values were replaced with 0
	if resultMap["notes"] != 0.0 {
		t.Errorf("Expected notes to be 0, got %v", resultMap["notes"])
	}

	// Check that numeric values were bounded
	meta := resultMap["meta"].(map[string]interface{})
	profile := meta["profile"].(map[string]interface{})
	if profile["id"] != 100.0 {
		t.Errorf("Expected id to be bounded to 100, got %v", profile["id"])
	}

	// Check that strings with numbers were replaced
	if resultMap["MIX"] != "REDACTED" {
		t.Errorf("Expected MIX to be REDACTED, got %v", resultMap["MIX"])
	}
}

func TestFilteringWithTransformations(t *testing.T) {
	input := createTestInput()

	transforms := &Transformations{
		BoundNum: &BoundRule{Min: 10, Max: 100},
	}
	filters := &Filters{
		MinKeyLen:  4,
		NoValTypes: []string{"null"},
		MaxDepth:   999999,
		MaxStrLen:  999999,
	}

	result := processJSON(input, filters, transforms, 1)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	// Check that short keys were filtered out
	if _, exists := resultMap["SYM"]; exists {
		t.Error("Expected SYM key to be filtered out (length < 4)")
	}

	if _, exists := resultMap["arr"]; exists {
		t.Error("Expected arr key to be filtered out (length < 4)")
	}

	// "zero" has length 4, so it should exist
	if _, exists := resultMap["zero"]; !exists {
		t.Error("Expected zero key to exist (length = 4)")
	} else if resultMap["zero"] != 10.0 {
		t.Errorf("Expected zero to be bounded to 10, got %v", resultMap["zero"])
	}

	// Check that nested null values were filtered, but transformed values remain
	meta := resultMap["meta"].(map[string]interface{})
	if meta == nil {
		t.Fatal("Expected meta to exist")
	}

	profile := meta["profile"].(map[string]interface{})
	if profile == nil {
		t.Fatal("Expected profile to exist")
	}

	// notes should be filtered out because it's null
	if _, exists := profile["notes"]; exists {
		t.Error("Expected notes to be filtered out (null value)")
	}
}

// Tests for command-line compatibility
func TestFullWorkflow(t *testing.T) {
	input := createTestInput()
	inputFile := "test_input.json"
	outputFile := "test_output.json"

	// Write test input
	if err := writeJSONFile(inputFile, input); err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	// Test masking workflow
	transforms := &Transformations{
		MaskVal: []MaskRule{
			{Pattern: "email", Mask: "***MASKED***"},
			{Pattern: "Name", Mask: "MASK"},
		},
	}
	filters := &Filters{MaxDepth: 999999, MaxKeyLen: 999999, MaxStrLen: 999999}

	result := processJSON(input, filters, transforms, 1)

	// Write result
	if err := writeJSONFile(outputFile, result); err != nil {
		t.Fatalf("Failed to write output file: %v", err)
	}

	// Read and verify
	output, err := readJSONFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputMap := output.(map[string]interface{})
	if outputMap["email"] != "***MASKED***" {
		t.Errorf("Expected email to be masked, got %v", outputMap["email"])
	}

	if outputMap["Name"] != "MASK" {
		t.Errorf("Expected Name to be masked, got %v", outputMap["Name"])
	}
}

// Benchmark tests
func BenchmarkProcessLargeJSON(b *testing.B) {
	// Create a larger test input
	largeInput := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		largeInput[fmt.Sprintf("key_%d", i)] = map[string]interface{}{
			"value":  float64(i),
			"text":   fmt.Sprintf("text_%d", i),
			"nested": createTestInput(),
		}
	}

	transforms := &Transformations{
		BoundNum: &BoundRule{Min: 0, Max: 500},
		ReplaceVal: []ReplaceRule{
			{Pattern: "upper", Replacement: "MASKED"},
		},
	}
	filters := &Filters{MaxDepth: 999999, MaxKeyLen: 999999, MaxStrLen: 999999}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processJSON(largeInput, filters, transforms, 1)
	}
}

// Run all tests
func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
