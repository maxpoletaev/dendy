package input

import "github.com/maxpoletaev/dendy/internal/binario"

type Device interface {
	Read() uint8
	Write(uint8)
	Reset()
	SaveState(w *binario.Writer) error
	LoadState(r *binario.Reader) error
}
