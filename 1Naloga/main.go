package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/laspp/PS-2023/vaje/naloga-1/koda/xkcd"
)

func main() {
	numWorkers := 32

	wordCounts := make(map[string]int)
	var mu sync.Mutex // Mutex za dostop do wordCounts

	comicChan := make(chan int)

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(comicChan, wordCounts, &mu, &wg)
	}

	// Pridobimo število stripov
	totalComics, err := getTotalComics()
	if err != nil {
		log.Fatalf("Napaka pri pridobivanju števila stripov: %v", err)
	}

	for i := 1; i <= totalComics; i++ {
		comicChan <- i
	}
	close(comicChan)

	wg.Wait()

	wordCountPairs := sortWordCounts(wordCounts)

	// Izpis 15 najpogostejših besed
	//fmt.Println("15 most frequent words:")
	for i := 0; i < 15 && i < len(wordCountPairs); i++ {
		fmt.Printf("%s: %d\n", wordCountPairs[i].Word, wordCountPairs[i].Count)
	}
}

// Delavec (gorutina guy)
func worker(comicChan <-chan int, wordCounts map[string]int, mu *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	for id := range comicChan {
		comic, err := xkcd.FetchComic(id)
		if err != nil {
			log.Printf("Napaka pri pridobivanju stripa %d: %v", id, err)
			continue
		}

		text := preprocessComic(comic)
		countWords(text, wordCounts, mu)
	}
}

// Predprocesiranje besedila
func preprocessComic(comic xkcd.Comic) string {
	var text string
	if comic.Transcript != "" {
		text = comic.Transcript
	} else {
		text = comic.Title + " " + comic.Tooltip
	}

	text = strings.ToLower(text)
	text = strings.Map(func(r rune) rune {
		if strings.ContainsRune("abcdefghijklmnopqrstuvwxyz0123456789 ", r) {
			return r
		}
		return -1
	}, text)

	return text
}

// Štetje besed
func countWords(text string, wordCounts map[string]int, mu *sync.Mutex) {
	words := strings.Fields(text)
	mu.Lock()
	defer mu.Unlock()
	for _, word := range words {
		if len(word) >= 4 {
			wordCounts[word]++
		}
	}
}

// Struct za shranjevanje besed in frekvenc
type wordCountPair struct {
	Word  string
	Count int
}

// Funkcija za sortiranje
func sortWordCounts(wordCounts map[string]int) []wordCountPair {
	var wordCountPairs []wordCountPair
	for word, count := range wordCounts {
		wordCountPairs = append(wordCountPairs, wordCountPair{Word: word, Count: count})
	}

	sort.Slice(wordCountPairs, func(i, j int) bool {
		return wordCountPairs[i].Count > wordCountPairs[j].Count
	})

	return wordCountPairs
}

// Stevilo stripov
func getTotalComics() (int, error) {
	latestComic, err := xkcd.FetchComic(0)
	if err != nil {
		return 0, err
	}

	return latestComic.Id, nil
}
