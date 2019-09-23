package core

func isNegative(n uint8) bool {
	return (n & 0b10000000) != 0
}

func toAbs(n uint8) uint8 {
	return n & 0b01111111
}
