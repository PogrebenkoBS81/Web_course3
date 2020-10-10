package server

import (
	"bufio"
	"context"
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
	host string
	port int
}

// NewServer - returns a new websocket server.
func NewServer(host string, port, interval int) *Server {
	return &Server{
		TimeManager : newTimeManager(interval),
		ClientManager: newClientManager(),
		// It's possible to pass it via CLI and validate it,
		// but there was nothing about protocol type it in lab,
		// so i don't want to experiment
		protocol: "tcp",
		host: host,
		port: port,
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
		if err := listener.Close(); err  != nil {
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
func (s *Server) handleConnections(ctx context.Context, listener net.Listener)  {
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
		case c := <- conns:
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
func (s *Server) updater(ctx context.Context)  {
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
func (s *Server) waitForClient(listener net.Listener, conns chan<-net.Conn) {
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
	req, err := s.readRequest(conn)
	if err != nil {
		log.Println("ERROR: error reading request, ", err)
		return
	}

	ready := make(chan bool, 1)
	hash := s.addClient(req.ClientName, conn.RemoteAddr().String(), ready)
	defer s.cleanClient(conn, hash)

	s.processClient(ctx, conn, ready)
}

// processClient - processes client connection
func (s *Server) processClient(ctx context.Context, conn net.Conn, ready<-chan bool) {
	// don't want to recreate writer over and over again inside the "writeToSocket"
	connWriter :=  bufio.NewWriter(conn)

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down...")

			return
		case <- ready:
			if err := s.writeToSocket(connWriter, s.getClients()); err != nil {
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

// writeToSocket - writes data to socket
func (s *Server) writeToSocket(writer *bufio.Writer, bts []byte) error {
	_, err := writer.Write(bts)
	if  err != nil {
		return err
	}

	return writer.Flush() // maybe, conn.Write() would be enough, but i'd better be sure.
}

// readRequest - reads request from client
func (s *Server) readRequest(conn net.Conn) (*request, error){
	// Only one message from client would be sent,
	// so create reader right here.
	reader := bufio.NewReader(conn)

	// Reader would be empty until peek.
	if _, err := reader.Peek(1); err != nil {
		return nil, err
	}
	buff := make([]byte, reader.Buffered())

	_, err := reader.Read(buff)
	if err != nil {
		return nil, err
	}

	resp := new(request)

	return resp, xml.Unmarshal(buff, resp)
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