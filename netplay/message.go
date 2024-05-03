package netplay

import (
	"errors"

	"github.com/maxpoletaev/dendy/internal/binario"
	"github.com/maxpoletaev/dendy/internal/bytepool"
)

type MsgType = uint8

const (
	MsgTypeReset MsgType = iota + 1
	MsgTypeWait
	MsgTypeInput
	MsgTypePing
	MsgTypePong
	MsgTypeBye
)

type Message struct {
	Type       MsgType
	Frame      uint32
	Generation uint32
	Payload    bytepool.Buffer
}

func readMsg(r *binario.Reader, msg *Message, pool *bytepool.Pool) error {
	if err := errors.Join(
		r.ReadUint8To(&msg.Type),
		r.ReadUint32To(&msg.Frame),
		r.ReadUint32To(&msg.Generation),
	); err != nil {
		return err
	}

	size, err := r.ReadUint32()
	if err != nil {
		return err
	}

	if size > 0 {
		msg.Payload = pool.Get(int(size))
		if err := r.ReadRawBytesTo(msg.Payload.Data); err != nil {
			return err
		}
	}

	return nil
}

func writeMsg(w *binario.Writer, msg *Message) error {
	defer msg.Payload.Free()

	return errors.Join(
		w.WriteUint8(msg.Type),
		w.WriteUint32(msg.Frame),
		w.WriteUint32(msg.Generation),
		w.WriteUint32(uint32(len(msg.Payload.Data))),
		w.WriteRawBytes(msg.Payload.Data),
	)
}
