package mappers

type Mapper interface {
	Initialize(prgRomBanks uint8, chrRomBanks uint8, prgMem []uint8, chrMem []uint8)

	Read(busId string, addr uint16, debug bool) (uint8, bool)
	Write(busId string, addr uint16, data uint8, debug bool) bool
}
