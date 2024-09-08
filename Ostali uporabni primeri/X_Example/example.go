package main

import (
	"fmt"
)

func TestOutput(*t Testing) string {
	n_goroutines := 10

	// Pricakovani izhodi
    expected = {
        "niz1 \n niz2 \n niz3 \n ...",
        "niz1 \n niz2 \n niz3 \n ...",
        ...
    }
    
	wg := sync.WaitGroup{}

	// Zagon procesov in zajem izhodov
    outputs = make ([]string, len(n_goroutines))
    for i := range n_goroutines {
		wg.Add(1)
        outputs[i] := go runProcessAndCaptureOutput(...)
    }

	wg.Wait()

	// Preverjanje izhoda procesa s njegovim pricakovanim izhodom
	for i := range n_goroutines {
		if outputs[i] != expected[i] {
			return fmt.Sprintf("Expected %s, got %s", expected[i], outputs[i])
		}
	}
}