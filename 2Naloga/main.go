package main

import (
	"flag"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"naloga2/socialNetwork"
)

// InvertedIndex is a structure representing an inverted index.
type InvertedIndex struct {
	mu    sync.RWMutex
	index map[string][]uint64
}

// NewInvertedIndex initializes a new InvertedIndex.
func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		index: make(map[string][]uint64),
	}
}

// Add adds words to the inverted index for a given post ID.
func (ii *InvertedIndex) Add(id uint64, words []string) {
	ii.mu.Lock()
	defer ii.mu.Unlock()
	for _, word := range words {
		if _, ok := ii.index[word]; !ok {
			ii.index[word] = make([]uint64, 0)
		}
		if !contains(ii.index[word], id) {
			ii.index[word] = append(ii.index[word], id)
		}
	}
}

// contains checks if a given ID is already present in the list.
func contains(ids []uint64, id uint64) bool {
	for _, existingID := range ids {
		if existingID == id {
			return true
		}
	}
	return false
}

// Search searches for all post IDs that contain the given word.
func (ii *InvertedIndex) Search(word string) []uint64 {
	ii.mu.RLock()
	defer ii.mu.RUnlock()
	return ii.index[word]
}

// formatWords processes the text by removing non-alphabetic characters and converting it to lowercase.
func formatWords(data string) string {
	re := regexp.MustCompile(`[^a-zA-Z]`)
	words := re.ReplaceAllString(data, " ")
	return strings.ToLower(words)
}

// formatWords2 processes the text by removing non-alphabetic characters and converting it to lowercase without spaces.
func formatWords2(data string) string {
	re := regexp.MustCompile(`[^a-zA-Z]`)
	return strings.ToLower(re.ReplaceAllString(data, ""))
}

// extractKeywords extracts keywords from the given text.
func extractKeywords(data string) []string {
	words := formatWords(data)
	wordsSplit := strings.Split(words, " ")

	var keywords []string
	for _, word := range wordsSplit {
		if len(word) >= 4 {
			keywords = append(keywords, word)
		}
	}
	return keywords
}

// worker processes tasks from high and low priority channels.
func worker(highPriorityChan, lowPriorityChan <-chan socialNetwork.Task, invertedIndex *InvertedIndex, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case task, ok := <-highPriorityChan:
			if !ok {
				return
			}
			if task.TaskType == "index" {
				keywords := extractKeywords(task.Data)
				invertedIndex.Add(task.Id, keywords)
				fmt.Printf("Indexed post with ID: %d\n", task.Id)
			}
		default:
			select {
			case task, ok := <-lowPriorityChan:
				if !ok {
					return
				}
				if task.TaskType == "search" {
					word := formatWords2(task.Data)
					results := invertedIndex.Search(word)
					fmt.Printf("Search results for word '%s': %v\n", word, results)
				}
			default:

			}
		}
	}
}

// main is the entry point of the program.
func main() {
	// Command-line flags
	numWorkers := flag.Int("workers", 5, "Number of workers")
	rateLimit := flag.Int("rate", 10, "Rate limit (requests per second)")
	flag.Parse()

	// Start time measurement
	start := time.Now()

	// Initialize the task generator `Q`.
	var producer socialNetwork.Q
	producer.New(0.5)

	// Create channels for high and low priority tasks.
	highPriorityChan := make(chan socialNetwork.Task)
	lowPriorityChan := make(chan socialNetwork.Task)
	invertedIndex := NewInvertedIndex()

	// Set up rate limiting
	var rateLimiter <-chan time.Time
	if *rateLimit > 0 {
		rateLimiter = time.Tick(time.Second / time.Duration(*rateLimit))
	}

	// Launch workers to process tasks.
	var wg sync.WaitGroup
	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go worker(highPriorityChan, lowPriorityChan, invertedIndex, &wg)
	}

	// Distribute tasks to appropriate channels based on rate limiting
	go func() {
		for {
			select {
			case <-rateLimiter:
				select {
				case task, ok := <-producer.TaskChan:
					if !ok {
						return
					}
					if task.TaskType == "index" {
						select {
						case highPriorityChan <- task:
						default:
						}
					} else {
						select {
						case lowPriorityChan <- task:
						default:
						}
					}
				default:
				}
			default:
				if rateLimit == nil {
					task, ok := <-producer.TaskChan
					if !ok {
						return
					}
					if task.TaskType == "index" {
						select {
						case highPriorityChan <- task:
						default:
						}
					} else {
						select {
						case lowPriorityChan <- task:
						default:
						}
					}
				}
			}
		}
	}()

	// Run the task generator and wait for a certain time
	go producer.Run()
	time.Sleep(time.Second * 10)
	producer.Stop()

	// Close channels and wait for workers to finish
	close(highPriorityChan)
	close(lowPriorityChan)
	wg.Wait()

	// Print the number of processed requests per second
	elapsed := time.Since(start)
	fmt.Printf("Processed requests rate: %f reqs/s\n", float64(producer.N[socialNetwork.LowPriority]+producer.N[socialNetwork.HighPriority])/float64(elapsed.Seconds()))
}
