// Package network is ...
package network

import (
    "fmt"
    "net"
    "log"
    "os"
    "io"
)

// Whatever...
type Channels struct {
    Ctrl        chan string
    DataIn      chan * DataChunk
    DataOut     chan * DataChunk
    Done        chan bool
    Connected   chan bool
    Panic       chan bool  
}

// Blah-blah...
type Connection struct {
    Channels
    Server              bool
    Address             string
    Name                string
    logger              * log.Logger
    socketDisconnect    chan bool
    conn                net.Conn
}

// Thingamajig...
type DataChunk struct {
    Length  int
    Bytes   []byte
}

// Start, Init, End...all wrapped in one, efficient, no!?
func (c * Connection) Start() {
    c.logger = log.New(os.Stdout, fmt.Sprintf("[%s] ", c.Name), log.Lmicroseconds)

    defer func() {
        if r:= recover(); r!= nil {
            c.logger.Println("Recovering...")
            c.Panic <- true
        }
    } ()

    if c.Server {
        ln, _ := net.Listen("tcp", ":" + c.Address)
        c.logger.Println("Listening on port", c.Address)
        c.conn, _ = ln.Accept()
    } else {
        var err error
        c.logger.Println("Trying to connect to", c.Address)
        c.conn, err = net.Dial("tcp", c.Address)
        c.logger.Println(err)
    }

    c.logger.Println("Connected to", c.conn.RemoteAddr())
    c.Connected <- true

    c.socketDisconnect = make(chan bool)

    go func() {
        for {
            select {
            case ctrl := <- c.Ctrl:
                if ctrl == "quit" {
                    c.logger.Println("Quitting receive goroutine")
                    c.Connected <- false
                    return
                }
            default:
                buffer := make([]byte, 1024)
                numBytes, err := c.conn.Read(buffer)

                if err == nil {
                    c.DataIn <- &DataChunk{Length: numBytes, Bytes: buffer[:numBytes]}
                } else {
                    if err == io.EOF {
                        c.socketDisconnect <- true
                        return
                    }
                }
            }
        }
        
    } ()

    for {
        select {
        case ctrl := <- c.Ctrl:
            if ctrl == "quit" {
                c.logger.Println("Quitting send goroutine")
                c.Done <- true
                return
            }
        case dataOut := <- c.DataOut:
            fmt.Println(string(dataOut.Bytes))
        case <- c.socketDisconnect:
            c.logger.Println("Socket disconnected")
            c.Connected <- false
            return
        }
    }
}
