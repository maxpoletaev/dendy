package relay

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/maxpoletaev/dendy/internal/binario"
)

func send(conn net.Conn, msg Message) error {
	writer := binario.NewWriter(conn, binary.LittleEndian)

	var typ MsgType
	switch msg.(type) {
	case *CreateSessionMsg:
		typ = MsgTypeCreateSession
	case *SessionCreatedMsg:
		typ = MsgTypeSessionCreated
	case *JoinSessionMsg:
		typ = MsgTypeJoinSession
	case *StartGameMsg:
		typ = MsgTypeStartGame
	case *ErrorMsg:
		typ = MsgTypeError
	default:
		return fmt.Errorf("unknown message type: %T", msg)
	}

	if err := writer.WriteUint8(typ); err != nil {
		return fmt.Errorf("failed to write message type: %w", err)
	}

	if err := msg.ToBytes(writer); err != nil {
		return fmt.Errorf("failed to write message body: %w", err)
	}

	return nil
}

func sendKeepAlive(conn net.Conn) error {
	w := binario.NewWriter(conn, binary.LittleEndian)
	return w.WriteUint8(MsgTypeKeepAlive)
}

func sendError(conn net.Conn, errToSend error) {
	writer := binario.NewWriter(conn, binary.LittleEndian)
	msg := &ErrorMsg{Message: errToSend.Error()}

	err := errors.Join(
		writer.WriteUint8(MsgTypeError),
		msg.ToBytes(writer),
	)

	if err != nil {
		log.Printf("[ERROR] failed to send error message: %s", err)
	}
}

func receive(r io.Reader) (msg Message, _ error) {
	reader := binario.NewReader(r, binary.LittleEndian)

retry:
	typ, err := reader.ReadUint8()
	if err != nil {
		return msg, fmt.Errorf("failed to read message type: %w", err)
	}

	switch typ {
	case MsgTypeCreateSession:
		msg = &CreateSessionMsg{}
	case MsgTypeSessionCreated:
		msg = &SessionCreatedMsg{}
	case MsgTypeJoinSession:
		msg = &JoinSessionMsg{}
	case MsgTypeStartGame:
		msg = &StartGameMsg{}
	case MsgTypeError:
		msg = &ErrorMsg{}
	case MsgTypeKeepAlive:
		goto retry
	default:
		return msg, fmt.Errorf("unexpected message type: %d", typ)
	}

	if err = msg.FromBytes(reader); err != nil {
		return msg, fmt.Errorf("failed to read message body: %w", err)
	}

	return msg, nil
}

func receiveType[T Message](conn net.Conn) (msg T, err error) {
	res, err := receive(conn)
	if err != nil {
		return msg, err
	}

	switch m := res.(type) {
	case T:
		return m, nil
	case *ErrorMsg:
		return msg, m
	default:
		return msg, fmt.Errorf("unexpected message type: %T", m)
	}
}
