package server

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

// ClientManager - manages the incoming clients.
// I calculate clientsXML once per tick,
// to avoid recalculations and remarshalling the same data for each client
// It's bad too, anyway. Marshalling data by each routine is fine, as far as i now.
type ClientManager struct {
	clients    map[string]*clientData // hash is used to avoid collisions.
	clientsXML []byte
	mutex      *sync.Mutex
}

// ClientData - manages the incoming clients.
type clientData struct {
	ready chan<- bool
	name  string
	ip    string
	time  int64
}

// newClientManager - returns a new client manager.
func newClientManager() *ClientManager {
	return &ClientManager{
		clientsXML: nil,
		clients:    make(map[string]*clientData),
		mutex:      &sync.Mutex{},
	}
}

// addClient - adds new client to the map by the generated hash.
func (c *ClientManager) addClient(clientName, clientAddr string, ch chan<- bool) string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Println("Adding new client:", clientName)

	// too lazy to use UUID here, hash is fine.
	keyString := strconv.FormatInt(time.Now().Unix(), 10) + clientName + clientAddr
	uniqueKey := md5.Sum([]byte(keyString))
	hash := hex.EncodeToString(uniqueKey[:])

	c.clients[hash] = &clientData{
		ready: ch,
		name:  clientName,
		ip:    clientAddr,
		time:  time.Now().Unix(),
	}

	return hash
}

// delClient - removes client from map.
func (c *ClientManager) delClient(clientHash string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	val := c.clients[clientHash]
	if val == nil {
		return errors.New(fmt.Sprintf("client with hash %s dosen't exists.", clientHash))
	}

	log.Println("Removing client:", val.name)

	delete(c.clients, clientHash)

	return nil
}

// getClients - returns XML list of clients in bytes.
func (c *ClientManager) getClients() []byte {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.clientsXML
}

// notifyClients - notify each routine about new data to send.
func (c *ClientManager) notifyClients() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Println("Notifying clients...")

	for _, client := range c.clients {
		client.ready <- true
	}
}

// notifyClients - updates XMl data on tick.
func (c *ClientManager) updateClients(time int64) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Println("Updating clients...")

	clientsList := make([]*client, len(c.clients))
	idx := 0

	for _, d := range c.clients {
		clientsList[idx] = &client{
			Connected: d.time,
			Name:      d.name,
			IP:        d.ip,
		}

		idx++
	}

	c.clientsXML, err = xml.Marshal(&response{
		Clients: clientsList,
		Timer:   time,
	})

	return
}
