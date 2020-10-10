package main

import (
	"lab2Client/cli"
	"lab2Client/client"
	"log"
	"strconv"
	"sync"
)

// runClients - runs specified amount of clients.
// It's superstructure upon client package.
func runClients(host string, port, clientsNum int) {
	var wg sync.WaitGroup // want to wait until all clients would be done
	wg.Add(clientsNum)

	for i := 0; i < clientsNum ;i++ {
		name := "Pogrebenko_" + strconv.Itoa(i)
		go clientRunner(&wg, host, name, port)
	}

	wg.Wait() // waiting for all clients to finish
}

// clientRunner - runs single client.
func clientRunner(wg *sync.WaitGroup, host, name string, port int) {
	defer wg.Done()

	if err := client.NewClient(host, port).Connect(name); err != nil{
		log.Println(err)
	}
}

func main() {
	config := cli.ParseArgs()
	log.SetFlags(log.LstdFlags|log.Lshortfile|log.Lmicroseconds)

	runClients(config.Host, config.Port, config.Clients)
}
