package mappers

type Mapper interface {
	Initialize(prgRomSize uint8, chrRomSize uint8, prgMem []uint8, chrMem []uint8)

	Read(addr uint16) (uint8, bool)
	Write(addr uint16, data uint8) bool

	PPURead(addr uint16) (uint8, bool)
	PPUWrite(addr uint16, data uint8) bool
}
