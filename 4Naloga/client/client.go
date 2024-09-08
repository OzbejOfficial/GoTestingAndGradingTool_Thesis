package main

import (
	"flag"
	"fmt"
	"net/rpc"
	"os"
	"strings"
	"time"

	"naloga4/storage"
)

// connectToServer vzpostavi povezavo z določenim strežnikom
func connectToServer(targetPort int) *rpc.Client {
	client, err := rpc.Dial("tcp", fmt.Sprintf("localhost:%d", targetPort))
	if err != nil {
		fmt.Println("Napaka pri povezavi na strežnik:", err)
		os.Exit(1)
	}
	return client
}

// handlePut zahteva PUT operacijo na strežniku
func handlePut(client *rpc.Client, task string, completed bool) {
	todo := storage.Todo{Task: task, Completed: completed}
	var reply string
	err := client.Call("RPCServer.Put", &todo, &reply)
	if err != nil {
		fmt.Println("Put [ERROR]:", task, "-", err, "<>", time.Now())
	} else {
		fmt.Println("Put:", todo.Task, "-", todo.Completed, "<>", time.Now())
	}
}

// handleGet zahteva GET operacijo na strežniku
func handleGet(client *rpc.Client, task string) {
	var result storage.Todo
	err := client.Call("RPCServer.Get", &task, &result)
	if err != nil {
		fmt.Println("Get [ERROR]:", task, "-", err, "<>", time.Now())
	} else {
		fmt.Println("Get:", result.Task, "-", result.Completed, "<>", time.Now())
	}
}

// main funkcija
func main() {
	port := flag.Int("port", 9000, "začetni port")
	n := flag.Int("n", 2, "število procesov v verigi")

	// Seznami: "get,put,put,get,get,put", "hello,world,science,hello,hello,science", "true,false,true,false,true,false"

	actionFlag := flag.String("action", "get", "seznam akcij, ki jih izvedemo (get ali put)")
	taskFlag := flag.String("task", "hello world", "seznam nalog")
	completedFlag := flag.String("completed", "true", "seznam ali je naloga dokončana (velja le za put)")

	flag.Parse()

	actionList := strings.Split(*actionFlag, ",")
	taskList := strings.Split(*taskFlag, ",")
	completedList := strings.Split(*completedFlag, ",")

	if len(actionList) != len(taskList) || len(actionList) != len(completedList) {
		fmt.Println("Napaka: Seznami akcij, nalog in dokončanih nalog se morajo ujemati po dolžini.")
		os.Exit(1)
	}

	for i := 0; i < len(actionList); i++ {
		action := actionList[i]
		task := taskList[i]
		completed := completedList[i] == "true"

		// Nastavitev glave ali repa glede na akcijo
		var targetPort int
		if action == "put" {
			// Glava verige (prvi proces)
			targetPort = *port
		} else if action == "get" {
			// Rep verige (zadnji proces)
			targetPort = *port + *n - 1
		} else {
			fmt.Println("Napaka: Neveljavna akcija. Uporabite 'get' ali 'put'.")
			os.Exit(1)
		}

		client := connectToServer(targetPort)
		defer client.Close()

		if action == "put" {
			handlePut(client, task, completed)
		} else if action == "get" {
			handleGet(client, task)
		}
	}
}
