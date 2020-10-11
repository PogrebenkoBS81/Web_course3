package client

import (
	"bufio"
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

// Client - simple websocket client.
type Client struct {
	protocol string
	host     string
	port     int
}

// NewClient - returns a new websocket client.
func NewClient(host string, port int) *Client {
	return &Client{
		protocol: "tcp",
		host:     host,
		port:     port,
	}
}

// Connect - connects to the socket server.
func (c *Client) Connect(ID string) error {
	path := fmt.Sprintf("%s:%d", c.host, c.port)

	conn, err := net.Dial(c.protocol, path)
	if err != nil {
		return err
	}
	// I usually use helper function, but don't want to call it only once.
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println(err)
		}
	}() // dont want to loose possible error

	log.Println("Connected to", path)

	// send base data to the server
	if err := c.send(conn, ID); err != nil {
		return err
	}

	// Due to i want to achieve graceful shutdown,
	// message waiting was moved to another goroutine.
	// Otherwise, ctx.Done() will be processed only after some message arrives
	// (conn.Read() is a blocking operation)
	dataCh := make(chan *response)
	ctx, cancel := c.handleCancel()
	go c.messageChecker(conn, cancel, dataCh)

	return c.waitForMessage(ctx, dataCh, ID)
}

// waitForMessage - waits for incoming data to process.
func (c *Client) waitForMessage(ctx context.Context, data <-chan *response, ID string) error {
	// No ping is required, if connection was lost,
	// messageChecker will receive an error, and ctx.Done will be called.
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down...")

			return nil
		case d := <-data:
			c.prettyPrint(ID, d)
		}
	}
}

// send - sends message to the server.
func (c *Client) send(conn net.Conn, clientID string) error {
	log.Println("Sending id:", clientID)
	req := &request{
		ClientName: clientID,
	}

	bts, err := xml.Marshal(req)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(conn)

	_, err = writer.Write(bts)
	if err != nil {
		return err
	}

	return writer.Flush() // maybe, conn.Write() would be enough, but i'd better be sure.
}

// messageChecker - waits for message
func (c *Client) messageChecker(conn net.Conn, cancel context.CancelFunc, data chan<- *response) {
	connReader := bufio.NewReader(conn) // don't want to recreate reader over and over again

	for {
		resp, err := c.readResponse(connReader)
		if err != nil {
			log.Println(err)
			cancel() // ctx.Done() if there is an error
			return
		}

		data <- resp
	}
}

// prettyPrint - pretty prints the server response.
func (c *Client) prettyPrint(clientId string, resp *response) {
	if len(resp.Clients) == 0 {
		return
	}

	message := clientId + " received:"
	for _, c := range resp.Clients {
		message += fmt.Sprintf("\n Client: %s; IP: %s, Connected: %d; Timer: %d;",
			c.Name, c.IP, c.Connected, resp.Timer)
	}

	log.Println(message)
}

// readResponse - reads server response.
func (c *Client) readResponse(reader *bufio.Reader) (*response, error) {
	// Reader would be empty until peek.
	if _, err := reader.Peek(1); err != nil {
		return nil, err
	}

	// Create buffer with required size
	buff := make([]byte, reader.Buffered())

	_, err := reader.Read(buff)
	if err != nil {
		return nil, err
	}

	resp := new(response)

	return resp, xml.Unmarshal(buff, resp)
}

// handleCancel - handles cancellation.
func (c *Client) handleCancel() (context.Context, context.CancelFunc) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go c.cancelHelper(signals, cancel)

	return ctx, cancel
}

// cancelHelper - calls cancel() function when signal arrived.
func (c *Client) cancelHelper(signals chan os.Signal, cancel context.CancelFunc) {
	sig := <-signals
	log.Println("Incoming signal:", sig)
	cancel()
}
