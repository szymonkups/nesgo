package core

type Controller struct {
	buttons [8]bool
	index   byte
	strobe  byte
}

type button uint8

const (
	ButtonA button = iota
	ButtonB
	ButtonSelect
	ButtonStart
	ButtonUp
	ButtonDown
	ButtonLeft
	ButtonRight
)

func (c *Controller) Read(_ string, addr uint16, _ bool) (uint8, bool) {
	if addr == 0x4016 {
		value := byte(0)
		if c.index < 8 && c.buttons[c.index] {
			value = 1
		}
		c.index++
		if c.strobe&1 == 1 {
			c.index = 0
		}
		return value, true
	}

	return 0x00, false
}

func (c *Controller) Write(_ string, addr uint16, data uint8, _ bool) bool {
	if addr == 0x4016 {
		c.strobe = data

		if c.strobe&1 == 1 {
			c.index = 0
		}

		return true
	}

	return false
}

func (c *Controller) PressButton(button button) {
	c.buttons[button] = true
}

func (c *Controller) ReleaseButton(button button) {
	c.buttons[button] = false
}
