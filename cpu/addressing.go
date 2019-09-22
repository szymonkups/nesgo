package cpu

// Addressing mode is a function that returns true if there is a potantial that
// it might require additional clock cycle
type addressingMode func() bool

func immediateAddr() bool {
	return false
}

func zeroPageAddr() bool {
	return false
}

func zeroPageXAddr() bool {
	return false
}

func absoluteAddr() bool {
	return false
}

func absoluteXAddr() bool {
	return false
}

func absoluteYAddr() bool {
	return false
}

func indirectXAddr() bool {
	return false
}

func indirectYAddr() bool {
	return false
}

func impliedAddr() bool {
	return false
}
