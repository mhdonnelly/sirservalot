package bridge

import (
    "fmt"
    "io"
    "log"
    "net"
)

const channelBufSize = 100
const readBufferSize = 1024
var maxId int = 0

type TCPClient struct {
    id int
    conn net.Conn
    server *TCPServer
    ch chan []byte
    doneCh chan bool
}

func NewTCPClient(conn net.Conn, tcpServer *TCPServer) *TCPClient {
    if conn == nil {
        panic("Connection cannot be nil")
    }

    if tcpServer == nil {
        panic("TCP Server cannot be nil")
    }

    maxId++
    ch := make(chan []byte, channelBufSize)
    doneCh := make(chan bool)

    return &TCPClient{maxId, conn, tcpServer, ch, doneCh}
}

func (c *TCPClient) Conn() net.Conn {
    return c.conn
}

func (c *TCPClient) Write(msg []byte) {
    select {
    case c.ch <- msg:
    default:
        c.server.Del(c)
        err := fmt.Errorf("client %d is disconnected.", c.id)
        c.server.Err(err)
    }
}

func (c *TCPClient) Done() {
    c.doneCh <- true
}

func (c *TCPClient) listenWrite() {
    for {
        select {

        // send message to the client
        case msg := <-c.ch:
            c.conn.Write(msg)

        // receive done request
        case <-c.doneCh:
            c.server.Del(c)
            c.doneCh <- true // for listenRead method
            return
        }
    }
}

func (c *TCPClient) listenRead() {
    for {
        select {

        case <-c.doneCh:
            c.server.Del(c)
            c.doneCh <- true // for listenWrite method
            return

        default:
            readBuffer := make([]byte, readBufferSize)
            n, err := c.conn.Read(readBuffer)
            if err == io.EOF {
                c.doneCh <- true
            } else if err != nil {
                c.server.Err(err)
            } else {
                log.Printf("Read %d bytes from %s\n", n, c.conn.RemoteAddr())
                c.server.tcpReceived <- readBuffer[:n]
            }
        }
    }
}

func (c *TCPClient) Listen() {
    go c.listenWrite()
    c.listenRead()
}

