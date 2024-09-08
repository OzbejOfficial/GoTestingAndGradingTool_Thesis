package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

type LogEntry struct {
	Time       time.Time
	Action     string
	Task       string
	Completed  bool
	FromServer bool // True if from server, false if from client
	Found      bool // Relevant for GET actions from client
}

// TestReplicationChain preverja pravilno delovanje verige z replikacijo
func TestReplicationChain(t *testing.T) {
	port := 12345
	n := 3
	servers := []*exec.Cmd{}
	var wg sync.WaitGroup

	logFiles := make([]*os.File, n)
	for i := 0; i < n; i++ {
		logFile, err := os.Create(fmt.Sprintf("server%d_log.txt", i))
		if err != nil {
			t.Fatalf("Napaka pri ustvarjanju log datoteke za strežnik %d: %v", i, err)
		}
		logFiles[i] = logFile
		defer logFiles[i].Close()
	}

	wg.Add(n)
	for i := 0; i < n; i++ {
		go SetupServer(port, i, n, &wg, logFiles)
	}

	wg.Wait()

	// Dodaten čas, da se strežniki popolnoma inicializirajo
	time.Sleep(2 * time.Second)

	clientCmd := SetupClient(port, n,
		"get,put,get,put,get,put,put,get,put,put,get",
		"ninja,hello,hello,world,world,banana,banana,banana,pink,pink,pink",
		"_,true,_,false,_,true,false,_,false,true,_",
		"log_client.txt")

	err := clientCmd.Run()
	if err != nil {
		t.Fatalf("Napaka pri zagonu klienta: %v", err)
	}

	// Zberi in združi vse dogodke iz log datotek
	serverLogs := []string{"server2_log.txt"}
	clientLog := "log_client.txt"
	events, err := collectLogEvents(clientLog, serverLogs)
	if err != nil {
		t.Fatalf("Napaka pri zbiranju dogodkov iz log datotek: %v", err)
	}

	// Simuliraj stanje in preveri pravilnost GET dogodkov
	if err := simulateAndCheck(events); err != nil {
		t.Fatalf("Napaka pri simulaciji in preverjanju: %v", err)
	}

	for _, cmd := range servers {
		TearDownServer(cmd)
	}
}

// SetupServer zažene strežnik
func SetupServer(port, id, n int, wg *sync.WaitGroup, logFiles []*os.File) {
	defer wg.Done() // Oznaka, da je ta gorutina končana

	cmd := exec.Command("go", "run", "main.go",
		fmt.Sprintf("-port=%d", port),
		fmt.Sprintf("-id=%d", id),
		fmt.Sprintf("-n=%d", n))
	cmd.Stdout = logFiles[id]
	cmd.Stderr = logFiles[id]
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Napaka pri zagonu strežnika na portu %d: %v\n", port+id, err)
		os.Exit(1)
	}
}

// ustavi strežnik
func TearDownServer(cmd *exec.Cmd) {
	cmd.Process.Kill()
	cmd.Wait()
}

// SetupClient zažene klienta
func SetupClient(port, n int, action, task, completed, outputFile string) *exec.Cmd {
	cmd := exec.Command("go", "run", "client.go",
		fmt.Sprintf("-port=%d", port),
		fmt.Sprintf("-n=%d", n),
		fmt.Sprintf("-action=%s", action),
		fmt.Sprintf("-task=%s", task),
		fmt.Sprintf("-completed=%s", completed))

	cmd.Dir = "./client"
	logFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Napaka pri ustvarjanju log datoteke za klienta: %v\n", err)
		os.Exit(1)
	}
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	return cmd
}

// collectLogEvents zbira vse dogodke iz klienta in strežnikov
func collectLogEvents(clientLog string, serverLogs []string) ([]LogEntry, error) {
	var events []LogEntry

	// Preberi client log in dodaj GET dogodke
	clientEvents, err := parseClientLog(clientLog)
	if err != nil {
		return nil, fmt.Errorf("Napaka pri obdelavi klientovega loga: %v", err)
	}
	events = append(events, clientEvents...)

	// Preberi strežniške loge in dodaj PUT dogodke
	for _, serverLog := range serverLogs {
		serverEvents, err := parseServerLog(serverLog)
		if err != nil {
			return nil, fmt.Errorf("Napaka pri obdelavi strežniškega loga: %v", err)
		}
		events = append(events, serverEvents...)
	}

	// Razvrsti dogodke po času
	sort.Slice(events, func(i, j int) bool {
		return events[i].Time.Before(events[j].Time)
	})

	return events, nil
}

// parseClientLog obdeluje client log in zbira GET dogodke
func parseClientLog(logFile string) ([]LogEntry, error) {
	file, err := os.Open(logFile)
	if err != nil {
		return nil, fmt.Errorf("Napaka pri odpiranju log datoteke: %v", err)
	}
	defer file.Close()

	var events []LogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Get:") || strings.HasPrefix(line, "Get [ERROR]:") {
			entry, err := parseClientLogLine(line)
			if err != nil {
				return nil, err
			}
			events = append(events, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Napaka pri branju log datoteke: %v", err)
	}

	return events, nil
}

// parseServerLog obdeluje server log in zbira PUT dogodke
func parseServerLog(logFile string) ([]LogEntry, error) {
	file, err := os.Open(logFile)
	if err != nil {
		return nil, fmt.Errorf("Napaka pri odpiranju log datoteke: %v", err)
	}
	defer file.Close()

	var events []LogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Put:") {
			entry, err := parseServerLogLine(line)
			if err != nil {
				return nil, err
			}
			events = append(events, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Napaka pri branju log datoteke: %v", err)
	}

	return events, nil
}

// parseClientLogLine obdeluje eno vrstico iz klientovega loga
func parseClientLogLine(line string) (LogEntry, error) {
	parts := strings.Split(line, " <> ")
	if len(parts) < 2 {
		return LogEntry{}, fmt.Errorf("Nepravilna oblika vrstice: %s", line)
	}

	timeStr := normalizeTimestamp(strings.Split(parts[1], " m=")[0])

	timeStamp, err := time.Parse("2006-01-02 15:04:05.000000000 -0700 MST", timeStr)
	if err != nil {
		return LogEntry{}, fmt.Errorf("Napaka pri parsiranju časa: %v", err)
	}

	actionParts := strings.Fields(parts[0])
	task := actionParts[1]

	var completed bool
	var found bool
	if strings.HasPrefix(line, "Get:") {
		completed = actionParts[3] == "true"
		found = true
	} else {
		completed = false
		found = false
	}

	return LogEntry{
		Time:       timeStamp,
		Action:     "Get",
		Task:       task,
		Completed:  completed,
		FromServer: false,
		Found:      found,
	}, nil
}

// parseServerLogLine obdeluje eno vrstico iz strežniškega loga
func parseServerLogLine(line string) (LogEntry, error) {
	parts := strings.Split(line, " <> ")
	if len(parts) < 2 {
		return LogEntry{}, fmt.Errorf("Nepravilna oblika vrstice: %s", line)
	}

	timeStr := normalizeTimestamp(strings.Split(parts[1], " m=")[0])

	timeStamp, err := time.Parse("2006-01-02 15:04:05.000000000 -0700 MST", timeStr)
	if err != nil {
		return LogEntry{}, fmt.Errorf("Napaka pri parsiranju časa: %v", err)
	}

	actionParts := strings.Fields(parts[0])
	task := actionParts[1]
	completed := actionParts[3] == "true"

	return LogEntry{
		Time:       timeStamp,
		Action:     "Put",
		Task:       task,
		Completed:  completed,
		FromServer: true,
		Found:      true,
	}, nil
}

// simulateAndCheck simulira stanje in preverja GET zahteve
func simulateAndCheck(events []LogEntry) error {
	state := make(map[string]bool)

	for _, event := range events {
		if event.Action == "Put" && event.FromServer {
			state[event.Task] = event.Completed
		} else if event.Action == "Get" && !event.FromServer {
			if event.Found {
				expected, exists := state[event.Task]
				if !exists || expected != event.Completed {
					return fmt.Errorf("GET napaka za task '%s': pričakovano %v, dobili %v", event.Task, expected, event.Completed)
				}
			} else {
				if _, exists := state[event.Task]; exists {
					return fmt.Errorf("GET napaka za task '%s': pričakovano ni vpisov, dobili %v", event.Task, event.Completed)
				}
			}
		}
	}

	return nil
}

// normalizeTimestamp poskrbi, da ima časovni žig vedno 9 decimalnih mest
func normalizeTimestamp(timeStr string) string {
	parts := strings.Split(timeStr, ".")
	if len(parts) == 2 {
		decimalPart := parts[1]
		spaceIndex := strings.Index(decimalPart, " ")
		if spaceIndex != -1 {
			remaining := decimalPart[spaceIndex:]
			decimalPart = decimalPart[:spaceIndex]

			for len(decimalPart) < 9 {
				decimalPart += "0"
			}
			return parts[0] + "." + decimalPart + remaining
		}
	}
	return timeStr
}
