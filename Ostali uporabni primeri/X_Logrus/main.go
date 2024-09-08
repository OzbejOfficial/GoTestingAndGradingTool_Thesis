package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	// Nastavitev logiranja v datoteko
	file, err := os.OpenFile("program.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Nastavitev loggerja, ki zapisuje v datoteko
	log.SetOutput(file)

	// Nastavitev formata logiranja
	log.SetFormatter(&log.JSONFormatter{})

	log.WithFields(log.Fields{
		"a":   1,
		"b":   2,
		"sum": 3,
		"t":   time.Now(),
	}).Info("Klic funkcije CalculateSum")

	/*
		log.WithFields(log.Fields{
			"animal": "walrus",
			"number": 1,
			"size":   10,
			"t":      time.Now(),
		}).Info("A walrus appears")

		log.WithFields(log.Fields{
			"animal": "walrus",
			"number": 1,
			"size":   42,
			"t":      time.Now(),
		}).Warning("A bigger walrus appears")

		log.WithFields(log.Fields{
			"animal": "not a walrus",
			"number": nil,
			"size":   nil,
			"t":      time.Now(),
		}).Error("No walrus appeared")
	*/
}
