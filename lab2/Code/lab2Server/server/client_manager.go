package server

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"sync"
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

// addClient - adds new client to the map by the host - port value.
func (c *ClientManager) addClient(name, addr string, time int64, ch chan<- bool) string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Println("Adding new client:", name)
	// Since host:port value is unique, no collisions is possible.
	c.clients[addr] = &clientData{
		ready: ch,
		name:  name,
		ip:    addr,
		time:  time,
	}

	return addr
}

// delClient - removes client from map.
func (c *ClientManager) delClient(clientKey string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	val := c.clients[clientKey]
	if val == nil {
		return errors.New(fmt.Sprintf("client with key %s dosen't exists.", clientKey))
	}

	log.Println("Removing client:", val.name)

	delete(c.clients, clientKey)

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
