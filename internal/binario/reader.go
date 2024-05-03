package binario

import (
	"encoding/binary"
	"io"
)

type Reader struct {
	byteOrder binary.ByteOrder
	reader    io.Reader
	buf       [8]byte
}

func NewReader(reader io.Reader, byteOrder binary.ByteOrder) *Reader {
	return &Reader{
		reader:    reader,
		byteOrder: byteOrder,
	}
}

func (r *Reader) ReadBool() (bool, error) {
	b, err := r.ReadUint8()
	return b != 0, err
}

func (r *Reader) ReadBoolTo(dst *bool) error {
	b, err := r.ReadBool()
	if err != nil {
		return err
	}

	*dst = b
	return nil
}

func (r *Reader) ReadUint8() (uint8, error) {
	bs := r.buf[:1]
	if _, err := r.reader.Read(bs); err != nil {
		return 0, err
	}

	return bs[0], nil
}

func (r *Reader) ReadUint8To(dst *uint8) error {
	b, err := r.ReadUint8()
	if err != nil {
		return err
	}

	*dst = b
	return nil
}

func (r *Reader) ReadUint16() (uint16, error) {
	bs := r.buf[:2]
	if _, err := r.reader.Read(bs); err != nil {
		return 0, err
	}

	return r.byteOrder.Uint16(bs), nil
}

func (r *Reader) ReadUint16To(dst *uint16) error {
	b, err := r.ReadUint16()
	if err != nil {
		return err
	}

	*dst = b
	return nil
}

func (r *Reader) ReadUint32() (uint32, error) {
	bs := r.buf[:4]
	if _, err := r.reader.Read(bs); err != nil {
		return 0, err
	}

	return r.byteOrder.Uint32(bs), nil
}

func (r *Reader) ReadUint32To(dst *uint32) error {
	b, err := r.ReadUint32()
	if err != nil {
		return err
	}

	*dst = b
	return nil
}

func (r *Reader) ReadUint64() (uint64, error) {
	bs := r.buf[:8]
	if _, err := r.reader.Read(bs); err != nil {
		return 0, err
	}

	return r.byteOrder.Uint64(bs), nil
}

func (r *Reader) ReadUint64To(dst *uint64) error {
	b, err := r.ReadUint64()
	if err != nil {
		return err
	}

	*dst = b
	return nil
}

func (r *Reader) ReadByteSlice() ([]byte, error) {
	length, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}

	if length == 0 {
		return nil, nil
	}

	bs := make([]byte, length)
	if _, err = r.reader.Read(bs); err != nil {
		return nil, err
	}

	return bs, nil
}

func (r *Reader) ReadByteSliceTo(dst []byte) error {
	length, err := r.ReadUint32()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	if len(dst) < int(length) {
		return io.ErrShortBuffer
	}

	bs := dst[:length]
	if _, err = r.reader.Read(bs); err != nil {
		return err
	}

	return nil
}

func (r *Reader) ReadRawBytesTo(dst []byte) error {
	if _, err := r.reader.Read(dst); err != nil {
		return err
	}

	return nil
}

func (r *Reader) ReadString() (string, error) {
	bs, err := r.ReadByteSlice()
	return string(bs), err
}

func (r *Reader) ReadStringTo(dst *string) error {
	bs, err := r.ReadByteSlice()
	if err != nil {
		return err
	}

	*dst = string(bs)
	return nil
}

func (r *Reader) ReadVarUint() (uint64, error) {
	var value uint64
	var shift uint

	for {
		b, err := r.ReadUint8()
		if err != nil {
			return 0, err
		}

		value |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			break
		}

		shift += 7
	}

	return value, nil
}

func (r *Reader) ReadVarUintTo(dst *uint64) error {
	b, err := r.ReadVarUint()
	if err != nil {
		return err
	}

	*dst = b
	return nil
}
