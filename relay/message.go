package relay

import (
	"errors"
	"net"

	"github.com/maxpoletaev/dendy/internal/binario"
)

type MsgType = uint8

const (
	MsgTypeCreateSession MsgType = iota
	MsgTypeSessionCreated
	MsgTypeJoinSession
	MsgTypeStartGame
	MsgTypeKeepAlive
	MsgTypeError
)

type Message interface {
	ToBytes(w *binario.Writer) error
	FromBytes(r *binario.Reader) error
}

type HelloMsg struct{}

type CreateSessionMsg struct {
	RomCRC32 uint32
	Public   bool
}

func (m *CreateSessionMsg) ToBytes(w *binario.Writer) error {
	return errors.Join(
		w.WriteUint32(m.RomCRC32),
		w.WriteBool(m.Public),
	)
}

func (m *CreateSessionMsg) FromBytes(r *binario.Reader) error {
	return errors.Join(
		r.ReadUint32To(&m.RomCRC32),
		r.ReadBoolTo(&m.Public),
	)
}

type SessionCreatedMsg struct {
	ID string
}

func (m *SessionCreatedMsg) ToBytes(w *binario.Writer) error {
	return errors.Join(
		w.WriteString(m.ID),
	)
}

func (m *SessionCreatedMsg) FromBytes(r *binario.Reader) error {
	return errors.Join(
		r.ReadStringTo(&m.ID),
	)
}

type JoinSessionMsg struct {
	ID       string
	RomCRC32 uint32
}

func (m *JoinSessionMsg) ToBytes(w *binario.Writer) error {
	return errors.Join(
		w.WriteString(m.ID),
		w.WriteUint32(m.RomCRC32),
	)
}

func (m *JoinSessionMsg) FromBytes(r *binario.Reader) error {
	return errors.Join(
		r.ReadStringTo(&m.ID),
		r.ReadUint32To(&m.RomCRC32),
	)
}

type StartGameMsg struct {
	IP   net.IP
	Port uint16
}

func (m *StartGameMsg) ToBytes(w *binario.Writer) error {
	return errors.Join(
		w.WriteByteSlice(m.IP),
		w.WriteUint16(m.Port),
	)
}

func (m *StartGameMsg) FromBytes(r *binario.Reader) error {
	m.IP = make(net.IP, 4)

	return errors.Join(
		r.ReadByteSliceTo(m.IP),
		r.ReadUint16To(&m.Port),
	)
}

type ErrorMsg struct {
	Message string
}

func (m *ErrorMsg) Error() string {
	return m.Message
}

func (m *ErrorMsg) ToBytes(w *binario.Writer) error {
	return errors.Join(
		w.WriteString(m.Message),
	)
}

func (m *ErrorMsg) FromBytes(r *binario.Reader) error {
	return errors.Join(
		r.ReadStringTo(&m.Message),
	)
}
