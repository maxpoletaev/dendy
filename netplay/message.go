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
	Buffer     bytepool.Buffer
	Frame      uint32
	Generation uint32
	Type       MsgType
}

func readMsg(r *binario.Reader, msg *Message, pool *bytepool.BytePool) error {
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
		msg.Buffer = pool.Buffer(int(size))
		if err = r.ReadRawBytesTo(msg.Buffer.Data); err != nil {
			return err
		}
	}

	return nil
}

func writeMsg(w *binario.Writer, msg *Message) error {
	err := errors.Join(
		w.WriteUint8(msg.Type),
		w.WriteUint32(msg.Frame),
		w.WriteUint32(msg.Generation),
		w.WriteUint32(uint32(len(msg.Buffer.Data))),
		w.WriteRawBytes(msg.Buffer.Data),
	)

	msg.Buffer.Free()

	return err
}
