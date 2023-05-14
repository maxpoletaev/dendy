package input

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

func (c *Joystick) Read() byte {
	value := byte(0)
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
