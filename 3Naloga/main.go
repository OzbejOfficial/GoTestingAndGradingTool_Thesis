package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {
	id := flag.Int("id", 0, "Identifikator procesa")
	N := flag.Int("N", 1, "Število vseh procesov")
	M := flag.Int("M", 1, "Število sporočil, ki jih razširi glavni proces")
	K := flag.Int("K", 1, "Število procesov, katerim se posreduje sporočilo")
	basePort := flag.Int("baseport", 8000, "Osnovni port za procese")
	seedFlag := flag.String("randseed", "0", "Seed za generiranje naključnih števil")

	flag.Parse()

	seed := strings.Split(*seedFlag, ",")
	seedPtr := 0
	//fmt.Println("Seed:", seed)

	listenAddr := fmt.Sprintf("localhost:%d", *basePort+*id)
	conn, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		fmt.Println("Error creating listener:", err)
		return
	}
	defer conn.Close()

	receivedMessages := make(map[string]bool)

	// Funkcija za prejemanje sporočil
	go func() {
		buffer := make([]byte, 1024)
		for {
			conn.SetDeadline(time.Now().Add(5 * time.Second))
			n, _, err := conn.ReadFrom(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					break // timeout po 5 sekundah
				}
				fmt.Println("Error reading message:", err)
				return
			}
			message := string(buffer[:n])

			// Če smo že prejeli to sporočilo, ga ignoriramo
			if receivedMessages[message] {
				continue
			}
			receivedMessages[message] = true

			fmt.Printf("Process %d received message: %s\n", *id, message)

			// Pošiljanje nprej
			if *K == *N-1 { // nestrpno razširjanje
				for i := 0; i < *N; i++ {
					if i != *id {
						fmt.Println("Forwarding message", message, "to process", i)
						sendMessage(fmt.Sprintf("localhost:%d", *basePort+i), message)
					}
				}
			} else { // razširjanje z govoricami
				forwardToRandomProcesses(*N, *K, *id, *basePort, message, seed, &seedPtr)
			}
		}
	}()

	// Pošiljanje sporočil iz procesa z id == 0
	if *id == 0 {

		// Se prepričamo, da so vsi procesi pripravljeni
		time.Sleep(1 * time.Second)

		for i := 1; i <= *M; i++ {
			message := fmt.Sprintf("Message %d", i)

			// Označimo, da smo prejeli svoje sporočilo
			receivedMessages[message] = true
			//fmt.Println("Process 0 marked message as received:", message)

			forwardMessage(*N, *K, *id, *basePort, message, seed, &seedPtr)
			time.Sleep(2000 * time.Millisecond) // pavza med pošiljanji
		}
	}

	// Čakamo, da se zaključi prejemanje sporočil
	time.Sleep(10 * time.Second)
}

// Funkcija za pošiljanje sporočil vsem procesom
func forwardMessage(N, K, id, basePort int, message string, seed []string, seedPtr *int) {
	if K == N-1 {
		for i := 1; i < N; i++ {
			if i != id {
				fmt.Println("Forwarding message", message, "to process", i)
				sendMessage(fmt.Sprintf("localhost:%d", basePort+i), message)
			}
		}
	} else {
		forwardToRandomProcesses(N, K, id, basePort, message, seed, seedPtr)
	}
}

// Funkcija za pošiljanje sporočil naključnim procesom
func forwardToRandomProcesses(N, K, id, basePort int, message string, seed []string, seedPtr *int) {
	for i := 0; i < K; i++ {
		target := getRandomNumber(seed, seedPtr)
		fmt.Println("Forwarding message", message, "to process", target)
		if target == id {
			continue
		}
		sendMessage(fmt.Sprintf("localhost:%d", basePort+target), message)
	}

	/*
		for _, i := range targets {
			if i != id {
				sendMessage(fmt.Sprintf("localhost:%d", basePort+i), message)
			}
		}
	*/
}

// Funkcija za pošiljanje sporočil
func sendMessage(address, message string) {
	conn, err := net.Dial("udp", address)
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error writing message:", err)
	}

	//fmt.Printf("Sent message to %s: %s\n", address, message)
}

func getRandomNumber(seed []string, seedPtr *int) int {
	if *seedPtr >= len(seed) {
		*seedPtr = 0
	}
	num, _ := strconv.Atoi(seed[*seedPtr])
	*seedPtr++
	return num
}
