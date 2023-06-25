package netplay

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

type MsgType uint8

const (
	MsgTypeReset MsgType = iota + 1
	MsgTypeInput
	MsgTypeSleep
)

type Message struct {
	Type        MsgType
	Frame       int32
	Incarnation uint32
	Payload     []byte
}

func (m *Message) Encode() ([]byte, error) {
	var buf bytes.Buffer

	buf.Write([]byte{byte(m.Type)})
	err1 := binary.Write(&buf, binary.LittleEndian, m.Frame)
	err2 := binary.Write(&buf, binary.LittleEndian, m.Incarnation)
	err3 := binary.Write(&buf, binary.LittleEndian, uint32(len(m.Payload)))

	if err := errors.Join(err1, err2, err3); err != nil {
		return nil, fmt.Errorf("failed to encode message header: %v", err)
	}

	if _, err := buf.Write(m.Payload); err != nil {
		return nil, fmt.Errorf("failed to encode message payload: %v", err)
	}

	return buf.Bytes(), nil
}

func (m *Message) Decode(data []byte) error {
	buf := bytes.NewBuffer(data)
	m.Type = MsgType(buf.Next(1)[0])

	var payloadSize uint32
	err1 := binary.Read(buf, binary.LittleEndian, &m.Frame)
	err2 := binary.Read(buf, binary.LittleEndian, &m.Incarnation)
	err3 := binary.Read(buf, binary.LittleEndian, &payloadSize)

	if err := errors.Join(err1, err2, err3); err != nil {
		return fmt.Errorf("failed to decode message header: %v", err)
	}

	if payloadSize > 0 {
		m.Payload = make([]byte, payloadSize)
		if _, err := io.ReadFull(buf, m.Payload); err != nil {
			return fmt.Errorf("failed to decode message payload: %v", err)
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
