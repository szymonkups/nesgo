package flags

type Flags struct {
	data uint8
}

type flag uint8

const (
	// Carry
	C flag = 0
	// Zero
	Z flag = 1
	// Interrupt disable
	I flag = 2
	// Decimal mode
	D flag = 3
	// Overflow flag
	V flag = 6
	// Negative flag
	N flag = 7
)

func (fs *Flags) Set(f flag, v bool) {
	var flag uint8 = 1 << f

	if v {
		fs.data |= flag
	} else {
		fs.data &= ^flag
	}
}

func (fs *Flags) SetZN(v uint8) {
	fs.Set(Z, v == 0x00)
	fs.Set(N, v&0x80 != 0x00)
}

func (fs *Flags) Get(f flag) bool {
	var flag uint8 = 1 << f

	return (fs.data & flag) != 0
}

func (fs *Flags) GetByte() uint8 {
	return fs.data
}

func (fs *Flags) SetByte(d uint8) {
	fs.data = d
}
