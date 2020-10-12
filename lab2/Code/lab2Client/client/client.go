package client

import (
	"bufio"
	"context"
	"encoding/binary"
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
	}() // dont want to lose possible error

	log.Println("Connected to", path)

	// send base data to the server
	if err := c.write(conn, &request{ClientName: ID}); err != nil {
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

// write - sends message to the server.
func (c *Client) write(conn net.Conn, data interface{}) error {
	bts, err := xml.Marshal(data)
	if err != nil {
		return err
	}

	if err := c.writeFull(conn, bts); err != nil {
		return err
	}

	return nil
}

// writeFull - writes full message to socket (required due to reader size limitations).
func (c *Client) writeFull(conn net.Conn, bts []byte) error {
	size := len(bts)
	base := make([]byte, 4)
	binary.BigEndian.PutUint32(base, uint32(size))
	base = append(base, bts...)

	// Same as server.
	writer := bufio.NewWriterSize(conn, size)
	_, err := writer.Write(base)

	if err != nil {
		return err
	}

	// Flush data to be sure that all bts was sent.
	return writer.Flush()
}

// messageChecker - waits for message
func (c *Client) messageChecker(conn net.Conn, cancel context.CancelFunc, data chan<- *response) {
	connReader := bufio.NewReader(conn) // don't want to recreate reader over and over again

	for {
		resp := new(response)
		if err := c.read(connReader, resp); err != nil {
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

// read - reads server response.
func (c *Client) read(reader *bufio.Reader, data interface{}) error {
	// get int size from message
	size, err := c.getSize(reader)
	if err != nil {
		return err
	}

	bts, err := c.readFull(reader, int(size))
	if err != nil {
		return err
	}

	return xml.Unmarshal(bts, data)
}

// readFull - reads full message (required due to reader size limitations)
func (c *Client) readFull(reader *bufio.Reader, size int) ([]byte, error) {
	fullMsg := make([]byte, 0)

	// Without size in message, there is possible situation,
	// when message will be exactly 4096 bytes,
	// and Peek() will hang after read
	// + size allows to separate data is socket when there is multiple different messages
	for {
		buffSize := reader.Buffered() // max reader size == 4096

		// Get required chunk size
		chunkSize := 0
		if size < buffSize {
			chunkSize = size
		} else {
			chunkSize = buffSize
		}
		// Create tmp storage, read bytes into it, and append them to the full message.
		buff := make([]byte, chunkSize)

		_, err := reader.Read(buff)
		if err != nil {
			return nil, err
		}
		fullMsg = append(fullMsg, buff...)

		// Break if message is fully read.
		size -= chunkSize
		if size == 0 {
			break
		}

		// Reader would be empty until peek.
		if _, err := reader.Peek(1); err != nil {
			return nil, err
		}
	}

	return fullMsg, nil
}

// getSize - reads first 4 bytes of size
func (c *Client) getSize(reader *bufio.Reader) (uint32, error) {
	// Reader would be empty until peek.
	if _, err := reader.Peek(1); err != nil {
		return 0, err
	}

	btsSize := make([]byte, 4)

	_, err := reader.Read(btsSize)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(btsSize), nil
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
