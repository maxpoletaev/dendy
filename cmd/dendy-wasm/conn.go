package main

import (
	"fmt"
	"log"
	"net"
	"syscall/js"
	"time"
)

var (
	_ net.Conn = (*webrtcConn)(nil)
)

type webrtcConn struct {
	conn       js.Value
	localAddr  net.TCPAddr
	remoteAddr net.TCPAddr
	readChan   chan int
	buf        []byte
}

func newWebrtcConn(conn js.Value) *webrtcConn {
	buf := make([]byte, 1024*10)
	readChan := make(chan int)

	conn.Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) any {
		data := js.Global().Get("Uint8Array").New(args[0].Get("data"))
		dataLength := data.Length()
		if dataLength > len(buf) {
			panic(fmt.Sprintf("data length %d exceeds buffer size %d", dataLength, len(buf)))
		}
		js.CopyBytesToGo(buf, data)
		readChan <- dataLength
		return nil
	}))

	return &webrtcConn{
		buf:        buf,
		conn:       conn,
		readChan:   readChan,
		localAddr:  net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0},
		remoteAddr: net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0},
	}
}

func (c *webrtcConn) Read(p []byte) (n int, err error) {
	n = <-c.readChan
	copy(p, c.buf[:n])
	return n, nil
}

func (c *webrtcConn) Write(p []byte) (n int, err error) {
	data := js.Global().Get("Uint8Array").New(len(p))
	js.CopyBytesToJS(data, p)
	c.conn.Call("send", data)
	return len(p), nil
}

func (c *webrtcConn) LocalAddr() net.Addr {
	return &c.localAddr
}

func (c *webrtcConn) RemoteAddr() net.Addr {
	return &c.remoteAddr
}

func (c *webrtcConn) Close() error {
	log.Printf("[WARN] Close is not implemented for webrtcConn")
	return nil
}

func (c *webrtcConn) SetDeadline(t time.Time) error {
	log.Printf("[WARN] SetDeadline is not implemented for webrtcConn")
	return nil
}

func (c *webrtcConn) SetReadDeadline(t time.Time) error {
	log.Printf("[WARN] SetReadDeadline is not implemented for webrtcConn")
	return nil
}

func (c *webrtcConn) SetWriteDeadline(t time.Time) error {
	log.Printf("[WARN] SetWriteDeadline is not implemented for webrtcConn")
	return nil
}
