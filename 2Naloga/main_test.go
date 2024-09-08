package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func TestIndexing(t *testing.T) {
	invertedIndex := NewInvertedIndex()

	// Adding words to the index
	invertedIndex.Add(1, []string{"test", "example"})

	// Test if the word 'test' was added correctly
	if len(invertedIndex.Search("test")) != 1 {
		t.Errorf("Error: The word 'test' was not correctly added to the index.")
	}

	// Test if the word 'example' was added correctly
	if len(invertedIndex.Search("example")) != 1 {
		t.Errorf("Error: The word 'example' was not correctly added to the index.")
	}

	// Test if the word 'keyword' was added correctly
	if len(invertedIndex.Search("keyword")) != 0 {
		t.Errorf("Error: The word 'keyword' was not correctly added to the index.")
	}
}

func TestSearch(t *testing.T) {
	invertedIndex := NewInvertedIndex()

	// Adding words to the index
	invertedIndex.Add(1, []string{"test"})
	invertedIndex.Add(2, []string{"example"})

	// Test if the word 'test' is found correctly
	if len(invertedIndex.Search("test")) != 1 {
		t.Errorf("Error: The word 'test' was not found in the index.")
	}

	// Test if the word 'example' is found correctly
	if len(invertedIndex.Search("example")) != 1 {
		t.Errorf("Error: The word 'example' was not found in the index.")
	}
}

func TestKeywordExtraction(t *testing.T) {
	data := "Hello, this is a test message too! Testing keywords."
	expectedKeywords := []string{"hello", "this", "test", "message", "testing", "keywords"}

	keywords := extractKeywords(data)
	if len(keywords) != len(expectedKeywords) {
		t.Errorf("Error: The expected number of keywords does not match.")
	}

	for i, keyword := range keywords {
		if keyword != expectedKeywords[i] {
			t.Errorf("Error: The keyword '%s' does not match the expected '%s'.", keyword, expectedKeywords[i])
		}
	}
}

// runProcessAndCaptureOutput runs the Go program with the provided flags and captures its output.
func runProcessAndCaptureOutput(workers, rate int) (string, error) {
	cmd := exec.Command("go", "run", "main.go",
		"-workers", fmt.Sprintf("%d", workers),
		"-rate", fmt.Sprintf("%d", rate),
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start process: %v", err)
	}

	// Wait for the process to finish
	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("process finished with error: %v", err)
	}

	return out.String(), nil
}

// cleanString removes non-alphanumeric characters except for spaces, colons, and newlines.
func cleanString(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9:\n ]+`)
	return re.ReplaceAllString(s, "")
}

// TestProgramOutputWithFlags runs the program with specific flags and compares the output to expected results.
func TestProgramOutputWithFlags(t *testing.T) {
	// Run the process with the specified workers and rate
	actualOutput, err := runProcessAndCaptureOutput(32, 20)
	if err != nil {
		t.Fatalf("Failed to run process: %v", err)
	}

	// Read the expected output from results.txt
	expectedOutput, err := os.ReadFile("resitev.txt")
	if err != nil {
		t.Fatalf("Could not read results.txt: %v", err)
	}

	// Clean both actual and expected output
	actualOutput = cleanString(actualOutput)
	expectedOutputStr := cleanString(string(expectedOutput))

	expectedOutputList := strings.Split(expectedOutputStr, "\n")
	for i, line := range expectedOutputList {
		if !strings.Contains(actualOutput, line) {
			t.Errorf("Error: The output does not contain the expected line %d.", i+1)
		}
	}
}
