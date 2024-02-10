package relay

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	commandPing = "PING"
	commandPong = "PONG"
)

func readWithTimeout(conn *net.UDPConn, timeout time.Duration, buf []byte) (int, *net.UDPAddr, error) {
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return 0, nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read: %w", err)
	}

	return n, addr, nil
}

func writeWithTimeout(conn *net.UDPConn, addr *net.UDPAddr, timeout time.Duration, buf []byte) (int, error) {
	if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return 0, fmt.Errorf("failed to set write deadline: %w", err)
	}

	n, err := conn.WriteToUDP(buf, addr)
	if err != nil {
		return n, fmt.Errorf("failed to write: %w", err)
	}

	return n, nil
}

func HolePunchUDP(ctx context.Context, lAddr, rAddr *net.UDPAddr) error {
	log.Printf("[INFO] punching %s->%s", lAddr, rAddr)

	conn, err := net.ListenUDP("udp", lAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("[WARN] failed to close connection: %s", err)
		}
	}()

	readerCtx, cancelReader := context.WithCancel(ctx)
	writerCtx, cancelWriter := context.WithCancel(ctx)
	errg, ctx := errgroup.WithContext(ctx)
	writerDone := make(chan struct{})
	readerDone := make(chan struct{})

	// Reader
	errg.Go(func() error {
		buf := make([]byte, 4)
		defer close(readerDone)

		for {
			select {
			case <-readerCtx.Done():
				return readerCtx.Err()

			default:
				n, _, err := readWithTimeout(conn, 30*time.Second, buf)

				if err != nil {
					if readerCtx.Err() != nil {
						return readerCtx.Err()
					}

					return fmt.Errorf("failed to read from peer: %w", err)
				} else if n != len(buf) {
					return fmt.Errorf("unexpected message length: %d", n)
				}

				switch string(buf) {
				case commandPing:
					_, err := writeWithTimeout(
						conn,
						rAddr,
						30*time.Second,
						[]byte(commandPong),
					)

					if err != nil {
						if readerCtx.Err() != nil {
							return readerCtx.Err()
						}

						return fmt.Errorf("failed to write to peer: %w", err)
					}

				case commandPong:
					cancelWriter()

				default:
					return fmt.Errorf("unexpected message from peer")
				}
			}
		}
	})

	// Writer
	errg.Go(func() error {
		defer close(writerDone)

		for {
			sleep := 500 * time.Millisecond
			sleep += time.Duration(rand.Intn(500)) * time.Millisecond

			select {
			case <-writerCtx.Done():
				return writerCtx.Err()

			case <-time.After(sleep):
				_, err := writeWithTimeout(
					conn,
					rAddr,
					30*time.Second,
					[]byte(commandPing),
				)

				if err != nil {
					if writerCtx.Err() != nil {
						return writerCtx.Err()
					}

					return fmt.Errorf("failed to write to peer: %w", err)
				}
			}
		}
	})

	<-writerDone
	time.Sleep(1 * time.Second) // give the reader some time to finish
	_ = conn.SetDeadline(time.Now())
	cancelReader()

	if err := errg.Wait(); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}

		return err
	}

	log.Printf("[INFO] connection established")

	return nil
}
