//go:build !wasm

package netplay

import (
	"fmt"
	"net"

	"github.com/xtaci/kcp-go"
)

func Listen(protocol string, lAddr string, game *Game) (*Netplay, net.Addr, error) {
	switch protocol {
	case "tcp":
		return listenTCP(game, lAddr)
	case "udp":
		return listenUDP(game, lAddr)
	default:
		return nil, nil, fmt.Errorf("unknown protocol: %s", protocol)
	}
}

func Connect(protocol string, rAddr, lAddr string, game *Game) (*Netplay, net.Addr, error) {
	switch protocol {
	case "tcp":
		return connectTCP(game, rAddr)
	case "udp":
		return connectUDP(game, lAddr, rAddr)
	default:
		return nil, nil, fmt.Errorf("unknown protocol: %s", protocol)
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
	np.isHost = true
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
	np.isHost = true
	np.start()

	return np, conn.RemoteAddr(), nil
}

func connectTCP(game *Game, addr string) (*Netplay, net.Addr, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, nil, err
	}

	np := newNetplay(game, conn)
	np.start()

	return np, conn.RemoteAddr(), nil
}

func connectUDP(game *Game, lAddr, rAddr string) (*Netplay, net.Addr, error) {
	lAddrUDP, err := net.ResolveUDPAddr("udp", lAddr)
	if err != nil {
		return nil, nil, err
	}

	localConn, err := net.ListenUDP("udp", lAddrUDP)
	if err != nil {
		return nil, nil, err
	}

	conn, err := kcp.NewConn(rAddr, nil, 0, 0, localConn)
	if err != nil {
		return nil, nil, err
	}

	np := newNetplay(game, conn)
	np.start()

	return np, conn.RemoteAddr(), nil
}
