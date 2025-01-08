package netplay

import (
	"net"
)

func Listen(protocol string, lAddr string, game *Game) (*Netplay, net.Addr, error) {
	panic("not implemented")
}

func Connect(protocol string, rAddr, lAddr string, game *Game) (*Netplay, net.Addr, error) {
	panic("not implemented")
}
