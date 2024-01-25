package netplay

import (
	"fmt"
	"net"

	"github.com/xtaci/kcp-go"
)

type Protocol int

const (
	TCP Protocol = iota
	UDP
)

func Listen(game *Game, addr string, protocol Protocol) (*Netplay, net.Addr, error) {
	switch protocol {
	case TCP:
		return listenTCP(game, addr)
	case UDP:
		return listenUDP(game, addr)
	default:
		panic(fmt.Errorf("unknown protocol %d", protocol))
	}
}

func Connect(game *Game, addr string, protocol Protocol) (*Netplay, net.Addr, error) {
	switch protocol {
	case TCP:
		return connectTCP(game, addr)
	case UDP:
		return connectUDP(game, addr)
	default:
		panic(fmt.Errorf("unknown protocol %d", protocol))
	}
}

func listenTCP(game *Game, addr string) (*Netplay, net.Addr, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on %s: %v", addr, err)
	}

	conn, err := listener.Accept()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to accept connection: %v", err)
	}

	np := newNetplay(game, conn)
	np.start()

	return np, conn.RemoteAddr(), nil
}

func listenUDP(game *Game, addr string) (*Netplay, net.Addr, error) {
	listener, err := kcp.Listen(addr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on %s: %v", addr, err)
	}

	conn, err := listener.Accept()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to accept connection: %v", err)
	}

	np := newNetplay(game, conn)
	np.start()

	return np, conn.RemoteAddr(), nil
}

func connectTCP(game *Game, addr string) (*Netplay, net.Addr, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to %s: %v", addr, err)
	}

	np := newNetplay(game, conn)
	np.start()

	return np, conn.RemoteAddr(), nil
}

func connectUDP(game *Game, addr string) (*Netplay, net.Addr, error) {
	conn, err := kcp.Dial(addr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to %s: %v", addr, err)
	}

	np := newNetplay(game, conn)
	np.start()

	return np, conn.RemoteAddr(), nil
}
