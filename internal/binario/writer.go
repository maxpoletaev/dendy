package binario

import (
	"encoding/binary"
	"io"
)

// Writer is a wrapper around io.Writer that provides methods for writing binary
// data. It is similar to the binary.Write but avoids expensive type assertions
// by providing separate methods for each type.
type Writer struct {
	byteOrder binary.ByteOrder
	writer    io.Writer
	buf       [8]byte
}

// NewWriter returns a new Writer that writes to w using the specified byte order.
func NewWriter(w io.Writer, byteOrder binary.ByteOrder) *Writer {
	return &Writer{
		byteOrder: byteOrder,
		writer:    w,
	}
}

func (w *Writer) WriteBool(value bool) error {
	if value {
		return w.WriteUint8(1)
	}

	return w.WriteUint8(0)
}

// WriteUint8 writes a single byte.
func (w *Writer) WriteUint8(value uint8) error {
	bf := w.buf[:1]
	bf[0] = value
	_, err := w.writer.Write(bf)
	return err
}

// WriteUint16 writes a 16-bit unsigned integer.
func (w *Writer) WriteUint16(value uint16) error {
	bf := w.buf[:2]
	w.byteOrder.PutUint16(bf, value)
	_, err := w.writer.Write(bf)

	return err
}

// WriteUint32 writes a 32-bit unsigned integer.
func (w *Writer) WriteUint32(value uint32) error {
	bf := w.buf[:4]
	w.byteOrder.PutUint32(bf, value)
	_, err := w.writer.Write(bf)

	return err
}

// WriteUint64 writes a 64-bit unsigned integer.
func (w *Writer) WriteUint64(value uint64) error {
	bf := w.buf[:8]
	w.byteOrder.PutUint64(bf, value)
	_, err := w.writer.Write(bf)

	return err
}

// WriteByteSlice writes a byte slice prefixed with its length.
func (w *Writer) WriteByteSlice(value []byte) error {
	length := uint32(len(value))
	if err := w.WriteUint32(length); err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	if _, err := w.writer.Write(value); err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteRawBytes(value []byte) error {
	if _, err := w.writer.Write(value); err != nil {
		return err
	}

	return nil
}

// WriteString writes a string prefixed with its length.
func (w *Writer) WriteString(value string) error {
	return w.WriteByteSlice([]byte(value))
}

// WriteVarUint writes a variable-length encoded unsigned integer.
// See https://developers.google.com/protocol-buffers/docs/encoding#varints
func (w *Writer) WriteVarUint(value uint64) error {
	for value >= 0x80 {
		if err := w.WriteUint8(uint8(value) | 0x80); err != nil {
			return err
		}

		value >>= 7
	}

	return w.WriteUint8(uint8(value))
}
