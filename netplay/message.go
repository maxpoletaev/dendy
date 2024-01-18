package netplay

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/maxpoletaev/dendy/internal/binario"
)

type MsgType = uint8

const (
	MsgTypeReset MsgType = iota + 1
	MsgTypeInput
	MsgTypePing
	MsgTypePong
	MsgTypeBye
)

type Message struct {
	Type       MsgType
	Frame      uint32
	Generation uint32
	Payload    []byte
}

func (m *Message) Encode() ([]byte, error) {
	buf := bytes.Buffer{}
	w := binario.NewWriter(&buf, binary.LittleEndian)

	err := errors.Join(
		w.WriteUint8(m.Type),
		w.WriteUint32(m.Frame),
		w.WriteUint32(m.Generation),
		w.WriteBytes(m.Payload),
	)

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *Message) Decode(data []byte) error {
	buf := bytes.NewReader(data)
	reader := binario.NewReader(buf, binary.LittleEndian)

	err := errors.Join(
		reader.ReadUint8To(&m.Type),
		reader.ReadUint32To(&m.Frame),
		reader.ReadUint32To(&m.Generation),
	)

	if err != nil {
		return err
	}

	m.Payload, err = reader.ReadBytes()
	if err != nil {
		return err
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
