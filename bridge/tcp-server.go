package bridge

import (
    "log"
    "net"
    "os"
)

type TCPServer struct {
    bindAddress string
    clients map[int]*TCPClient
    addCh chan *TCPClient
    delCh chan *TCPClient
    doneCh chan bool
    errCh chan error
    serialReceived chan []byte
    tcpReceived chan []byte
}

func NewTCPServer(bindAddress string, serialReceived chan []byte, tcpReceived chan []byte) *TCPServer {
    return &TCPServer {
        bindAddress,
        make(map[int]*TCPClient),
        make(chan *TCPClient),
        make(chan *TCPClient),
        make(chan bool),
        make(chan error),
        serialReceived,
        tcpReceived,
    }
}

func (s *TCPServer) Add(c *TCPClient) {
    s.addCh <- c
}

func (s *TCPServer) Del(c *TCPClient) {
    s.delCh <- c
}

func (s *TCPServer) SendAll(msg []byte) {
    s.tcpReceived <- msg
}

func (s *TCPServer) Done() {
    s.doneCh <- true
}

func (s *TCPServer) Err(err error) {
    s.errCh <- err
}

func (s *TCPServer) sendAll(msg []byte) {
    data := append(msg, []byte("\n")...)
    for _, c := range s.clients {
        c.Write(data)
    }
}

func (s *TCPServer) acceptConnections() {
    onConnected := func(conn net.Conn) {
        defer func() {
            err := conn.Close()
            if err != nil {
                s.errCh <- err
            }
        }()

        // limit to 10 clients
        if len(s.clients) < 10 {
            client := NewTCPClient(conn, s)
            s.Add(client)
            client.Listen()
        }
    }

    l, err := net.Listen("tcp", s.bindAddress)
    if err != nil {
        log.Println("Error listening:", err.Error())
        os.Exit(1)
    }

    defer l.Close()

    for {
        conn, err := l.Accept()
        if err != nil {
            log.Println("Error accepting:", err.Error())
            os.Exit(1)
        }
        go onConnected(conn)
    }
}

func (s *TCPServer) Listen() {
    go s.acceptConnections()

    for {
        select {

        // Add new a client
        case c := <-s.addCh:
            s.clients[c.id] = c
            log.Printf("New client accepted, id: %d, total clients %d\n", c.id, len(s.clients))

        // del a client
        case c := <-s.delCh:
            delete(s.clients, c.id)
            log.Printf("Client disconnected, id: %d, total clients %d\n", c.id, len(s.clients))

        case msg := <-s.serialReceived:
            log.Printf("Writing %d bytes to %d clients\n", len(msg), len(s.clients))
            s.sendAll(msg)

        case err := <-s.errCh:
            log.Println(err.Error())

        case <-s.doneCh:
            return
        }
    }
}
