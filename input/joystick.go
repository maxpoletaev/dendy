package input

import (
	"errors"

	"github.com/maxpoletaev/dendy/internal/binario"
)

type Button int

const (
	ButtonA      Button = 0
	ButtonB      Button = 1
	ButtonSelect Button = 2
	ButtonStart  Button = 3
	ButtonUp     Button = 4
	ButtonDown   Button = 5
	ButtonLeft   Button = 6
	ButtonRight  Button = 7
)

type Joystick struct {
	buttons uint8
	index   uint8
	reset   uint8
}

func NewJoystick() *Joystick {
	return &Joystick{}
}

func (c *Joystick) Reset() {
	c.buttons = 0
	c.index = 0
	c.reset = 0
}

func (c *Joystick) SetButtons(buttons uint8) {
	c.buttons = buttons
}

func (c *Joystick) Read() (value byte) {
	value = (c.buttons >> c.index) & 0x01
	c.index++

	if c.reset&0x01 == 1 {
		c.index = 0
	}

	return value
}

func (c *Joystick) Write(value byte) {
	c.reset = value

	if c.reset&0x01 == 1 {
		c.index = 0
	}
}

func (c *Joystick) SaveState(w *binario.Writer) error {
	return errors.Join(
		w.WriteUint8(c.buttons),
		w.WriteUint8(c.index),
		w.WriteUint8(c.reset),
	)
}

func (c *Joystick) LoadState(r *binario.Reader) error {
	return errors.Join(
		r.ReadUint8To(&c.buttons),
		r.ReadUint8To(&c.index),
		r.ReadUint8To(&c.reset),
	)
}
