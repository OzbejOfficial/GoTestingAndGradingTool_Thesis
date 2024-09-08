package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

func TestNestrpnoRazsirjanje(t *testing.T) {
	basePort := 9000
	numProcesses := 4
	messages := 1
	K := numProcesses - 1 // nestrpno razširjanje

	// Definicija seedov za vsakega od procesov
	seeds := []string{
		"0", "0", "0", "0",
	}

	// Pričakovane log datoteke za vsak proces
	expectedLogs := []string{
		"log0_1.txt",
		"log1_1.txt",
		"log2_1.txt",
		"log3_1.txt",
	}

	outputs := make([]string, numProcesses)
	var wg sync.WaitGroup

	// Zaženemo procese in zajamemo njihove izhode
	for i := 0; i < numProcesses; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			outputs[i] = runProcessAndCaptureOutput(i, numProcesses, messages, K, basePort, seeds[i])
		}(i)
	}

	// Počakamo, da vsi procesi zaključijo
	wg.Wait()

	// Preverimo, ali se izhod ujema s pričakovanim logom
	for i := 0; i < numProcesses; i++ {
		expectedOutput, err := os.ReadFile("resitev/" + expectedLogs[i])
		if err != nil {
			t.Errorf("Failed to read expected log file for process %d: %v", i, err)
			continue
		}

		if strings.TrimSpace(outputs[i]) != strings.TrimSpace(string(expectedOutput)) {
			t.Errorf("Output for process %d does not match expected output.\nGot:\n%s\nExpected:\n%s", i, outputs[i], string(expectedOutput))
		}
	}
}

func TestRazsirjanjeZGovoricami(t *testing.T) {
	basePort := 9100
	numProcesses := 4
	messages := 1
	K := 2 // razširjanje z govoricami

	// Definicija seedov za vsakega od procesov
	seeds := []string{
		"1,0,2,3",
		"2,1,3,0",
		"2,0,1,3",
		"0,1,2,3",
	}

	// Pričakovane log datoteke za vsak proces
	expectedLogs := []string{
		"log0_2.txt",
		"log1_2.txt",
		"log2_2.txt",
		"log3_2.txt",
	}

	outputs := make([]string, numProcesses)
	var wg sync.WaitGroup

	// Zaženemo procese in zajamemo njihove izhode
	for i := 0; i < numProcesses; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			outputs[i] = runProcessAndCaptureOutput(i, numProcesses, messages, K, basePort, seeds[i])
		}(i)
	}

	// Počakamo, da vsi procesi zaključijo
	wg.Wait()

	// Preverimo, ali se izhod ujema s pričakovanim logom
	for i := 0; i < numProcesses; i++ {
		expectedOutput, err := os.ReadFile("resitev/" + expectedLogs[i])
		if err != nil {
			t.Errorf("Failed to read expected log file for process %d: %v", i, err)
			continue
		}

		if strings.TrimSpace(outputs[i]) != strings.TrimSpace(string(expectedOutput)) {
			t.Errorf("Output for process %d does not match expected output.\nGot:\n%s\nExpected:\n%s", i, outputs[i], string(expectedOutput))
		}
	}
}

// Funkcija za zagon procesa in zajemanje njegovega izhoda
func runProcessAndCaptureOutput(id, N, M, K, basePort int, seed string) string {
	cmd := exec.Command("go", "run", "main.go",
		"-id", fmt.Sprintf("%d", id),
		"-N", fmt.Sprintf("%d", N),
		"-M", fmt.Sprintf("%d", M),
		"-K", fmt.Sprintf("%d", K),
		"-baseport", fmt.Sprintf("%d", basePort),
		"-randseed", seed,
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start process %d: %s\n", id, err)
		return ""
	}
	cmd.Wait()

	return out.String()
}
