package cpu

import (
	"encoding/gob"
	"errors"
	"fmt"
)

func (cpu *CPU) Save(enc *gob.Encoder) error {
	err := errors.Join(
		enc.Encode(cpu.A),
		enc.Encode(cpu.X),
		enc.Encode(cpu.Y),
		enc.Encode(cpu.P),
		enc.Encode(cpu.SP),
		enc.Encode(cpu.PC),
		enc.Encode(cpu.Cycles),
		enc.Encode(cpu.interrupt),
		enc.Encode(cpu.Halt),
	)

	if err != nil {
		return fmt.Errorf("failed to encode CPU state: %w", err)
	}

	return nil
}

func (cpu *CPU) Load(dec *gob.Decoder) error {
	err := errors.Join(
		dec.Decode(&cpu.A),
		dec.Decode(&cpu.X),
		dec.Decode(&cpu.Y),
		dec.Decode(&cpu.P),
		dec.Decode(&cpu.SP),
		dec.Decode(&cpu.PC),
		dec.Decode(&cpu.Cycles),
		dec.Decode(&cpu.interrupt),
		dec.Decode(&cpu.Halt),
	)

	if err != nil {
		return fmt.Errorf("failed to decode CPU state: %w", err)
	}

	return nil
}
