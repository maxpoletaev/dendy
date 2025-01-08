//go:build wasm

package netplay

import (
	"net"
)

func Listen(protocol string, lAddr string, game *Game) (*Netplay, net.Addr, error) {
	panic("not implemented for wasm")
}

func Connect(protocol string, rAddr, lAddr string, game *Game) (*Netplay, net.Addr, error) {
	panic("not implemented for wasm")
}
