package main

import (
	"log"
	"os"
)

func main() {
	// Ustvarjanje ali odpiranje log datoteke
	file, err := os.OpenFile("program.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Ustvarjanje loggerja, ki zapisuje v datoteko
	logger := log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	logger.Println("Program started")
	logger.Printf("Task: %s - Status: %s\n", "CalculateSum", "Success")
}
