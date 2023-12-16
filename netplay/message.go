package netplay

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

type MsgType = uint8

const (
	MsgTypeReset MsgType = iota + 1
	MsgTypeInput
	MsgTypePing
	MsgTypePong
)

type Message struct {
	Type       MsgType
	Frame      uint32
	Generation uint32
	Payload    []byte
}

func (m *Message) Encode() ([]byte, error) {
	var buf bytes.Buffer

	buf.Write([]byte{m.Type})
	err1 := binary.Write(&buf, binary.LittleEndian, m.Frame)
	err2 := binary.Write(&buf, binary.LittleEndian, m.Generation)
	err3 := binary.Write(&buf, binary.LittleEndian, uint32(len(m.Payload)))

	if err := errors.Join(err1, err2, err3); err != nil {
		return nil, err
	}

	if _, err := buf.Write(m.Payload); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *Message) Decode(data []byte) error {
	var (
		dataSize uint32
		err      error
	)

	buf := bytes.NewReader(data)
	m.Type, err = buf.ReadByte()
	if err != nil {
		return err
	}

	err1 := binary.Read(buf, binary.LittleEndian, &m.Frame)
	err2 := binary.Read(buf, binary.LittleEndian, &m.Generation)
	err3 := binary.Read(buf, binary.LittleEndian, &dataSize)

	if err = errors.Join(err1, err2, err3); err != nil {
		return err
	}

	if dataSize > 0 {
		m.Payload = make([]byte, dataSize)
		if _, err = io.ReadFull(buf, m.Payload); err != nil {
			return err
		}
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
		return msg, fmt.Errorf("failed to read message length: %w", err)
	}

	buf := make([]byte, size)
	if _, err = io.ReadFull(conn, buf); err != nil {
		return msg, fmt.Errorf("failed to read message: %w", err)
	}

	if err = msg.Decode(buf); err != nil {
		return msg, fmt.Errorf("failed to decode message: %w", err)
	}

	return msg, nil
}

func writeMsg(conn net.Conn, msg Message) error {
	data, err := msg.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	if err = binary.Write(conn, binary.LittleEndian, uint32(len(data))); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}

	if _, err = conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}
