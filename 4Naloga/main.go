package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"time"

	"naloga4/storage"
)

var (
	HEAD bool
	TAIL bool
)

type RPCServer struct {
	Storage *storage.TodoStorage
	Next    *rpc.Client
	Prev    *rpc.Client
	mu      sync.Mutex
}

func (s *RPCServer) Put(todo *storage.Todo, reply *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var ret struct{}
	if err := s.Storage.Put(todo, &ret); err != nil {
		return err
	} else {
		fmt.Println("Put:", todo.Task, "-", todo.Completed, "<>", time.Now())
	}
	if s.Next != nil {
		if err := s.Next.Call("RPCServer.Put", todo, reply); err != nil {
			return err
		}
	}
	/*
		} else {

				// Commit on tail in go routine but avoid sending put reply
				go func() {
					var commit_reply string
					if err := s.Commit(todo, &commit_reply); err != nil {
						fmt.Println("Error in commit:", err)
					}
				}()
			}
	*/
	*reply = "Put Successful"
	return nil
}

func (s *RPCServer) Commit(todo *storage.Todo, reply *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var ret struct{}
	if err := s.Storage.Commit(todo, &ret); err != nil {
		return err
	} else {
		fmt.Println("Commit:", todo.Task, "-", todo.Completed, "<>", time.Now())
	}
	if s.Prev != nil {
		if err := s.Prev.Call("RPCServer.Commit", todo, reply); err != nil {
			return err
		}
	}
	*reply = "Commit Successful"
	return nil
}

func (s *RPCServer) Get(task *string, result *storage.Todo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var dict = make(map[string]storage.Todo)
	todo := storage.Todo{Task: *task}
	if err := s.Storage.Get(&todo, &dict); err != nil {
		return err
	} else {
		fmt.Println("Get:", todo.Task, "-", todo.Completed, "<>", time.Now())
	}
	if val, exists := dict[*task]; exists { // && val.Commited
		*result = val
		return nil
	}
	return storage.ErrorNotFound
}

func main() {
	// Definicija vhodnih parametrov
	portPtr := flag.Int("port", 9000, "začetni port")
	idPtr := flag.Int("id", 0, "ID procesa")
	nPtr := flag.Int("n", 2, "število procesov v verigi")

	flag.Parse()

	port := *portPtr
	id := *idPtr
	N := *nPtr

	if id == 0 {
		HEAD = true
	}

	if id == N-1 {
		TAIL = true
	}

	serverAddr := fmt.Sprintf(":%d", port+id)
	keyValueStore := storage.NewTodoStorage()

	rpcServer := &RPCServer{Storage: keyValueStore}
	rpc.Register(rpcServer)

	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		fmt.Println("Napaka pri zagonu strežnika:", err)
		return
	}

	//fmt.Printf("Strežnik posluša na portu %d...\n", port+id)

	// Počakaj na zagon vseh strežnikov
	time.Sleep(1 * time.Second)

	// Nastavi Next in Prev
	if !TAIL {
		nextAddr := fmt.Sprintf("localhost:%d", port+id+1)
		next, err := rpc.Dial("tcp", nextAddr)
		if err != nil {
			fmt.Println("Napaka pri povezavi na naslednji strežnik:", err)
			return
		}
		rpcServer.Next = next
	}

	if !HEAD {
		prevAddr := fmt.Sprintf("localhost:%d", port+id-1)
		prev, err := rpc.Dial("tcp", prevAddr)
		if err != nil {
			fmt.Println("Napaka pri povezavi na prejšnji strežnik:", err)
			return
		}
		rpcServer.Prev = prev
	}

	// Poslušanje novih povezav
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Napaka pri sprejemu povezave:", err)
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()

	// Čas zagnanih strežnikov
	time.Sleep(10 * time.Second)
}
