package netplay

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"net"
)

type MsgType uint8

const (
	MsgTypeReset MsgType = iota + 1
	MsgTypeInput
)

type Message struct {
	Type    MsgType
	Frame   int32
	Payload []byte
}

func (m *Message) Encode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(m); err != nil {
		return nil, fmt.Errorf("failed to encode message: %v", err)
	}

	return buf.Bytes(), nil
}

func (m *Message) Decode(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(m); err != nil {
		return fmt.Errorf("failed to decode message: %v", err)
	}

	return nil
}

func readMsg(conn net.Conn) (Message, error) {
	var (
		msg Message
		err error
	)

	var size uint32
	if err = binary.Read(conn, binary.LittleEndian, &size); err != nil {
		return msg, fmt.Errorf("failed to read message length: %v", err)
	}

	buf := make([]byte, size)
	if _, err = io.ReadFull(conn, buf); err != nil {
		return msg, fmt.Errorf("failed to read message: %v", err)
	}

	if err = msg.Decode(buf); err != nil {
		return msg, fmt.Errorf("failed to decode message: %v", err)
	}

	return msg, nil
}

func writeMsg(conn net.Conn, msg Message) error {
	data, err := msg.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode message: %v", err)
	}

	if err = binary.Write(conn, binary.LittleEndian, uint32(len(data))); err != nil {
		return fmt.Errorf("failed to write message length: %v", err)
	}

	if _, err = conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	return nil
}
