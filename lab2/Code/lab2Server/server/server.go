package server

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	// it's fine for sync.Once to be global.
	// It's kind of a singleton pattern, and sometimes it really useful.
	doOnce sync.Once
)

// Server - simple websocket server.
type Server struct {
	*TimeManager
	*ClientManager
	protocol string
	host     string
	port     int
}

// NewServer - returns a new websocket server.
func NewServer(host string, port, interval int) *Server {
	return &Server{
		TimeManager:   newTimeManager(interval),
		ClientManager: newClientManager(),
		// It's possible to pass it via CLI and validate it,
		// but there was nothing about protocol type it in lab,
		// so i don't want to experiment
		protocol: "tcp",
		host:     host,
		port:     port,
	}
}

// Run - runs the websocket server.
func (s *Server) Run() error {
	path := fmt.Sprintf("%s:%d", s.host, s.port)
	log.Printf("Starting %s server on: %s", s.protocol, path)

	listener, err := net.Listen(s.protocol, path)
	if err != nil {
		return err
	}
	// I usually use helper function, but don't want to call it only once.
	defer func() {
		if err := listener.Close(); err != nil {
			log.Println(err)
		}
	}() // dont want to loose possible error

	// Pass context for the graceful shutdown.
	// It's not relevant in this lab (and it could be done far better),
	// but it's the best practice in go to use ctx
	s.handleConnections(s.handleCancel(), listener)

	return nil
}

// handleConnections - handles incoming connections
func (s *Server) handleConnections(ctx context.Context, listener net.Listener) {
	conns := make(chan net.Conn, 1)
	go s.waitForClient(listener, conns)

	for {
		select {
		// There is no need in context here,
		// but it's fine to pass ctx to functions +
		// in "possible" future there is ability to cancel some operations.
		case <-ctx.Done():
			log.Println("Exit call...")

			return
		case c := <-conns:
			// Start timer only once, when first successful connection is established.
			// (there was nothing in the task about stopping the timer, when there is no clients)
			doOnce.Do(func() {
				s.startTimer()
				go s.updater(ctx)
			})

			go s.handleClient(ctx, c)
		}
	}
}

// updater - updates client list and sends the notification to all routines to send it.
func (s *Server) updater(ctx context.Context) {
	for {
		select {
		// Same as handleConnections
		case <-ctx.Done():
			log.Println("Exit call...")

			return
		case <-s.ticker.C:
			if err := s.updateClients(s.getTime()); err == nil {
				s.notifyClients()
			} else {
				log.Println(err)
			}
		}
	}
}

// waitForClient - waits for the incoming connections
// if connection is successful - sends it to the connection handler
func (s *Server) waitForClient(listener net.Listener, conns chan<- net.Conn) {
	for {
		c, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("Client %s connected", c.RemoteAddr().String())
		conns <- c
	}

}

// handleClient - handles incoming client connection.
func (s *Server) handleClient(ctx context.Context, conn net.Conn) {
	reader := bufio.NewReader(conn)
	req := new(request)

	if err := s.read(reader, req); err != nil {
		log.Println("ERROR: error reading request, ", err)
		return
	}

	ready := make(chan bool, 1)
	hash := s.addClient(req.ClientName, conn.RemoteAddr().String(), ready)
	defer s.cleanClient(conn, hash)

	s.processClient(ctx, conn, ready)
}

// processClient - processes client connection
func (s *Server) processClient(ctx context.Context, conn net.Conn, ready <-chan bool) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down...")

			return
		case <-ready:
			if err := s.writeFull(conn, s.getClients()); err != nil {
				log.Println(err)

				return
			}
		default:
			// Ping 1 time in a second, so CPU won't be trashed by infinite loop.
			if !s.isAlive(conn) {
				return
			}
		}
	}
}

// Pings the connection. Waits for 1 second. If no EOF - connection is opened.
func (s *Server) isAlive(c net.Conn) bool {
	one := make([]byte, 1)
	if err := c.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		log.Println(err)
		return false
	}

	if _, err := c.Read(one); err == io.EOF {
		return false
	}

	return true
}

// cleanClient - cleans client data after it left.
func (s *Server) cleanClient(conn net.Conn, clientHash string) {
	log.Printf("Client %s, closing connection...", conn.RemoteAddr().String())
	if err := conn.Close(); err != nil {
		log.Println(err)
	}

	if err := s.delClient(clientHash); err != nil {
		log.Println("ERROR: error deleting client,", err)
	}
}

// writeFull - writes full message to socket (required due to reader size limitations).
func (s *Server) writeFull(conn net.Conn, bts []byte) error {
	size := len(bts)
	uintSize := uint32(size)

	base := make([]byte, 4)
	binary.BigEndian.PutUint32(base, uintSize)
	base = append(base, bts...)

	// maybe, it will be better to create 1 writer outside the function,
	// and write message the same way i read them,
	// but it's 3 AM here, i cant think, i'm already writing very bad code,
	// and i REALLY want to sleep.
	writer := bufio.NewWriterSize(conn, size)
	_, err := writer.Write(base)

	if err != nil {
		return err
	}

	// Flush data to be sure that all bts was sent.
	return writer.Flush()
}

// read - reads server response.
func (s *Server) read(reader *bufio.Reader, data interface{}) error {
	// get int size from message
	size, err := s.getSize(reader)
	if err != nil {
		return err
	}

	bts, err := s.readFull(reader, size)
	if err != nil {
		return err
	}

	return xml.Unmarshal(bts, data)
}

// readFull - reads full message (required due to reader size limitations)
func (s *Server) readFull(reader *bufio.Reader, msgSize uint32) ([]byte, error) {
	fullMsg := make([]byte, 0)
	size := int(msgSize)

	for  {
		tmpSize := 0 // size of the tmp storage for a part of the message
		buffSize := reader.Buffered() // max reader size == 4096

		// Find the required size
		if size > buffSize {
			tmpSize = buffSize
		} else {
			tmpSize = size
		}

		// Create tmp storage, read bytes into it, and append them to the full message.
		buff := make([]byte, tmpSize)

		_, err := reader.Read(buff)
		if err != nil {
			return nil, err
		}
		fullMsg = append(fullMsg, buff...)

		// Break if message is fully read.
		size -= tmpSize
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
func (s *Server) getSize(reader *bufio.Reader) (uint32, error) {
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

// handleCancel - handles cancellation
func (s *Server) handleCancel() context.Context {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go s.cancelHelper(signals, cancel)

	return ctx
}

// cancelHelper - calls cancel() function when signal arrived.
func (s *Server) cancelHelper(signals chan os.Signal, cancel context.CancelFunc) {
	sig := <-signals
	log.Println("Incoming signal:", sig)
	cancel()
}
