package input

type Button int

// auto increase
const (
	ButtonA Button = iota
	ButtonB
	ButtonSelect
	ButtonStart
	ButtonUp
	ButtonDown
	ButtonLeft
	ButtonRight
)

type Joystick struct {
	buttons [8]bool
	index   uint8
	reset   uint8
}

func NewJoystick() *Joystick {
	return &Joystick{}
}

func (c *Joystick) Press(button Button) {
	c.buttons[button] = true
}

func (c *Joystick) Release(button Button) {
	c.buttons[button] = false
}

func (c *Joystick) Read() (value byte) {
	if c.buttons[c.index] {
		value = 1
	}

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

func (c *Joystick) Buttons() [8]bool {
	return c.buttons
}
