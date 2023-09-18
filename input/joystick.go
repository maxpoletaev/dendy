package input

import (
	"encoding/gob"
	"errors"
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

func (c *Joystick) Save(enc *gob.Encoder) error {
	return errors.Join(
		enc.Encode(c.buttons),
		enc.Encode(c.index),
		enc.Encode(c.reset),
	)
}

func (c *Joystick) Load(dec *gob.Decoder) error {
	return errors.Join(
		dec.Decode(&c.buttons),
		dec.Decode(&c.index),
		dec.Decode(&c.reset),
	)
}
