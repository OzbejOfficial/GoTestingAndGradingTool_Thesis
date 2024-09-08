package main

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"regexp"
	"sync"
	"testing"

	"github.com/laspp/PS-2023/vaje/naloga-1/koda/xkcd"
)

// Test funkcije preprocessComic
func TestPreprocessComic(t *testing.T) {
	// Testni strip s podatki
	comic := xkcd.Comic{
		Title:      "Test Title",
		Transcript: "This is a transcript.",
		Tooltip:    "Tooltip should be ignored when Transcript is present.",
	}

	expected := "this is a transcript"
	result := preprocessComic(comic)

	if result != expected {
		t.Errorf("PreprocessComic() = %v, want %v", result, expected)
	}

	// Testni strip brez Transcript polja
	comic2 := xkcd.Comic{
		Title:   "Test Title",
		Tooltip: "This is a tooltip.",
	}

	expected2 := "test title this is a tooltip"
	result2 := preprocessComic(comic2)

	if result2 != expected2 {
		t.Errorf("PreprocessComic() = %v, want %v", result2, expected2)
	}
}

// Test funkcije countWords
func TestCountWords(t *testing.T) {
	wordCounts := make(map[string]int)
	text := "this is a test with test words like test and words"
	countWords(text, wordCounts, &sync.Mutex{})

	// Preverimo, če je štetje pravilno
	expectedCounts := map[string]int{
		"test":  3,
		"words": 2,
		"this":  1,
		"with":  1,
		"like":  1,
	}

	notExpectedCounts := map[string]int{
		"is":  1,
		"a":   1,
		"and": 1,
	}

	for word, expectedCount := range expectedCounts {
		if count, ok := wordCounts[word]; !ok || count != expectedCount {
			t.Errorf("countWords() = %v, want %v", count, expectedCount)
		}
	}

	// Preverimo, da besede s tremi ali manj znaki niso štete
	for word, count := range notExpectedCounts {
		if _, ok := wordCounts[word]; ok {
			t.Errorf("countWords() = %v, want %v", count, 0)
		}
	}
}

// Test funkcije za pravilno sortiranje in izpis
// Test funkcije za pravilno sortiranje besed po frekvenci
func TestSortWordCounts(t *testing.T) {
	// Testni podatki
	wordCounts := map[string]int{
		"test":    5,
		"words":   3,
		"example": 8,
		"go":      2,
		"code":    6,
	}

	// Pričakovan vrstni red po sortiranju
	expected := []wordCountPair{
		{Word: "example", Count: 8},
		{Word: "code", Count: 6},
		{Word: "test", Count: 5},
		{Word: "words", Count: 3},
		{Word: "go", Count: 2},
	}

	// Klic funkcije za sortiranje
	result := sortWordCounts(wordCounts)

	// Preverjanje rezultatov
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("sortWordCounts() = %v, want %v", result, expected)
	}
}

// Test funkcije getTotalComics
func TestGetTotalComics(t *testing.T) {
	// Pridobitev števila stripov s funkcijo getTotalComics
	totalComics, err := getTotalComics()
	if err != nil {
		t.Fatalf("getTotalComics() returned an error: %v", err)
	}

	// Pridobitev zadnjega stripa neposredno s FetchComic(0)
	latestComic, err := xkcd.FetchComic(0)
	if err != nil {
		t.Fatalf("FetchComic(0) returned an error: %v", err)
	}

	// Preverjanje, če je totalComics enak ID-ju zadnjega stripa
	if totalComics != latestComic.Id {
		t.Errorf("getTotalComics() = %v, want %v", totalComics, latestComic.Id)
	}
}

// cleanString removes non-alphanumeric characters except for spaces, colons, and newlines.
func cleanString(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9:\n ]+`)
	return re.ReplaceAllString(s, "")
}

// TestMainOutput tests the program's output against the expected output in results.txt.
func TestMainOutput(t *testing.T) {
	// Create a pipe to capture output
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Redirect standard output to the pipe
	stdout := os.Stdout
	os.Stdout = w

	// Goroutine to read from the pipe
	var buf bytes.Buffer
	done := make(chan bool)
	go func() {
		io.Copy(&buf, r)
		done <- true
	}()

	// Call the main function to generate output
	main()

	// Close the pipe and wait for the goroutine to finish
	w.Close()
	<-done

	// Restore standard output
	os.Stdout = stdout

	// Read the expected output from results.txt
	expectedOutput, err := os.ReadFile("resitev.txt")
	if err != nil {
		t.Fatalf("Could not read results.txt: %v", err)
	}

	// Clean both actual and expected output
	actualOutput := cleanString(buf.String())
	expectedOutputStr := cleanString(string(expectedOutput))

	// Compare the cleaned outputs
	if actualOutput != expectedOutputStr {
		t.Errorf("Output mismatch:\nGot:\n%s\nExpected:\n%s", actualOutput, expectedOutputStr)
	}
}
