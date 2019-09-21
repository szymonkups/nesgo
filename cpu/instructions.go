package cpu

// Increment X register by one
var inx = instruction{
	opCode:   0xE8,
	name:     "INX",
	noCycles: 2,
	addrMode: dummyAddressing,
	handler: func(cpu *CPU) bool {
		cpu.x++

		cpu.setFlag(z, cpu.x == 0x00)
		cpu.setFlag(n, cpu.x&0x80 != 0)

		return false
	},
}

// Increment Y register by one
var iny = instruction{
	opCode:   0xC8,
	name:     "INY",
	noCycles: 2,
	addrMode: dummyAddressing,
	handler: func(cpu *CPU) bool {
		cpu.y++

		cpu.setFlag(z, cpu.y == 0x00)
		cpu.setFlag(n, cpu.y&0x80 != 0)

		return false
	},
}
