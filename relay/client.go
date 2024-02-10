package relay

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xtaci/kcp-go"
)

type Client struct {
	wg        sync.WaitGroup
	stop      chan struct{}
	closed    atomic.Bool
	relayConn net.Conn
}

func Connect(relayAddr string) (*Client, error) {
	relayConn, err := kcp.Dial(relayAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial relay server: %w", err)
	}

	c := &Client{
		relayConn: relayConn,
		stop:      make(chan struct{}),
	}

	c.startKeepAlive()

	return c, nil
}

func (c *Client) startKeepAlive() {
	c.wg.Add(1)

	go func() {
		defer c.wg.Done()

		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-c.stop:
				return

			case <-ticker.C:
				if err := sendKeepAlive(c.relayConn); err != nil {
					if c.closed.Load() {
						return
					}

					log.Printf("[ERROR] failed to send keep alive: %s", err)
				}
			}
		}
	}()
}

func (c *Client) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}

	close(c.stop)

	c.wg.Wait()

	return c.relayConn.Close()
}

func (c *Client) CreateSession(romCRC32 uint32, public bool) (string, error) {
	err := send(c.relayConn, &CreateSessionMsg{
		RomCRC32: romCRC32,
		Public:   public,
	})

	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	res, err := receiveType[*SessionCreatedMsg](c.relayConn)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return res.ID, nil
}

func (c *Client) JoinSession(sessionID string, romCRC32 uint32) error {
	err := send(c.relayConn, &JoinSessionMsg{
		ID:       sessionID,
		RomCRC32: romCRC32,
	})

	if err != nil {
		return fmt.Errorf("failed to join session: %w", err)
	}

	return nil
}

func (c *Client) GetPeerAddress() (*net.UDPAddr, *net.UDPAddr, error) {
	msg, err := receiveType[*StartGameMsg](c.relayConn)
	if err != nil {
		return nil, nil, err
	}

	lAddr := c.relayConn.LocalAddr().(*net.UDPAddr)

	lAddr.IP = net.IPv4zero // bind to all interfaces

	rAddr := &net.UDPAddr{IP: msg.IP, Port: int(msg.Port)}

	return lAddr, rAddr, nil
}
