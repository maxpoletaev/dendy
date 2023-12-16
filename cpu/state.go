package cpu

import (
	"errors"

	"github.com/maxpoletaev/dendy/internal/binario"
)

func (cpu *CPU) SaveState(w *binario.Writer) error {
	return errors.Join(
		w.WriteUint8(cpu.A),
		w.WriteUint8(cpu.X),
		w.WriteUint8(cpu.Y),
		w.WriteUint8(cpu.P),
		w.WriteUint8(cpu.SP),
		w.WriteUint16(cpu.PC),
		w.WriteUint64(cpu.Cycles),
		w.WriteUint8(cpu.interrupt),
		w.WriteUint32(uint32(cpu.Halt)),
	)
}

func (cpu *CPU) LoadState(r *binario.Reader) error {
	var (
		halt uint32
	)

	err := errors.Join(
		r.ReadUint8To(&cpu.A),
		r.ReadUint8To(&cpu.X),
		r.ReadUint8To(&cpu.Y),
		r.ReadUint8To(&cpu.P),
		r.ReadUint8To(&cpu.SP),
		r.ReadUint16To(&cpu.PC),
		r.ReadUint64To(&cpu.Cycles),
		r.ReadUint8To(&cpu.interrupt),
		r.ReadUint32To(&halt),
	)

	cpu.Halt = int(halt)

	return err
}
