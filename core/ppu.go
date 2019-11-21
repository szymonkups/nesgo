package core

import (
	"fmt"
	"github.com/szymonkups/nesgo/core/ppu"
)

type PPU struct {
	NMI bool

	scanLine        int16
	cycle           int16
	isFrameComplete bool
	bus             *bus
	ctrlRegister    *ppu.ControlRegister
	statusRegister  *ppu.StatusRegister
	maskRegister    *ppu.MaskRegister
	addrRegister    *ppu.AddressRegister
	dataBuffer      uint8
}

func NewPPU(bus *bus) *PPU {
	newPPU := &PPU{
		scanLine:        -1,
		cycle:           0,
		isFrameComplete: false,
		bus:             bus,
		ctrlRegister:    new(ppu.ControlRegister),
		statusRegister:  new(ppu.StatusRegister),
		maskRegister:    new(ppu.MaskRegister),
		addrRegister:    new(ppu.AddressRegister),
		dataBuffer:      0,
	}

	newPPU.ctrlRegister.Write(0b00000000)
	newPPU.statusRegister.Write(0b00000000)
	newPPU.maskRegister.Write(0b00000000)

	// Set two times to zero it
	newPPU.addrRegister.Write(0b00000000)
	newPPU.addrRegister.Write(0b00000000)

	return newPPU
}

func (ppu *PPU) Read(_ string, addr uint16, debug bool) (uint8, bool) {
	if debug {
		panic(fmt.Errorf("debug read from ppu is not implemented yet"))
	}

	if addr >= 0x2000 && addr <= 0x3FFF {
		switch addr & 0x0007 {
		case 0x00:
			// PPU control register - PPUCTRL - write only
			return 0x00, true

		case 0x01:
			// PPU Mask Register 2 - PPUMASK - write only
			return 0x00, true

		case 0x02:
			ppu.statusRegister.SetVBlank(true)
			toReturn := (ppu.statusRegister.Read() & 0xE0) | (ppu.dataBuffer & 0x1F)
			ppu.statusRegister.SetVBlank(false)
			ppu.addrRegister.ResetLatch()
			return toReturn, true

		case 0x03:
			// Sprite Memory Address - OAMADDR - write only
			return 0x00, true

		case 0x04:
			// Sprite Memory Data - OAMDATA - read/write
			// TODO implement reading OAMDATA
			return 0x00, true

		case 0x05:
			// Screen Scroll Offset - PPUSCROLL - write only
			return 0x00, true

		case 0x06:
			// PPU Memory Address - PPUADDR - write only
			return 0x00, true

		case 0x07:
			address := ppu.addrRegister.GetAddress()

			// Get is normally delayed by 1 cycle...
			toReturn := ppu.dataBuffer
			ppu.dataBuffer = ppu.bus.Read(address)

			// ...until we read from palette memory
			if address >= 0x3f00 {
				toReturn = ppu.dataBuffer
			}

			ppu.addrRegister.Increment()
			return toReturn, true
		}
	}

	return 0x00, false
}

func (ppu *PPU) Write(_ string, addr uint16, data uint8, debug bool) bool {
	if debug {
		panic(fmt.Errorf("debug write to ppu is not implemented yet"))
	}

	// PPU registers exposed on CPU bus
	if addr >= 0x2000 && addr <= 0x3FFF {
		switch addr & 0x0007 {
		case 0x00:
			ppu.ctrlRegister.Write(data)
			return true

		case 0x01:
			ppu.maskRegister.Write(data)
			return true

		case 0x02:
			// PPU Status Register - PPUSTATUS - read only
			return true

		case 0x03:
			// Sprite Memory Address - OAMADDR - write only
			return true

		case 0x04:
			// Sprite Memory Data - OAMDATA - read/write
			// TODO implement reading OAMDATA
			return true

		case 0x05:
			// Screen Scroll Offset - PPUSCROLL - write only
			return true

		case 0x06:
			ppu.addrRegister.Write(data)
			return true

		case 0x07:
			ppu.bus.Write(ppu.addrRegister.GetAddress(), data)
			ppu.addrRegister.Increment()
			return true
		}
	}

	return false
}

func (ppu *PPU) Clock() {
	// Start VBlank
	if ppu.scanLine == 241 && ppu.cycle == 1 {
		ppu.statusRegister.SetVBlank(true)
		if ppu.ctrlRegister.NMIEnable {
			ppu.NMI = true
		}
	}

	// End VBlank
	if ppu.scanLine == -1 && ppu.cycle == 1 {
		ppu.statusRegister.SetVBlank(false)
	}

	ppu.cycle++
	// Single line of screen is 256 but scan line goes above that to 341
	if ppu.cycle >= 341 {
		ppu.cycle = 0
		ppu.scanLine++

		// We have 240 lines on screen but it goes above that to 261 (240 - 261 is called VBlank)
		if ppu.scanLine >= 261 {
			ppu.scanLine = -1
			ppu.isFrameComplete = true
		}
	}
}

func (ppu *PPU) GetPatternTables() {
}

type PPUColor struct {
	R uint8
	G uint8
	B uint8
}

func (ppu *PPU) GetUniversalBGColor() *PPUColor {
	i := ppu.bus.Read(0x3F00)

	if i >= 0x40 {
		fmt.Printf("Trying to access palette index out of range: %d.\n", i)
		return &PPUColor{0, 0, 0}
	}

	return colorTable[i]
}

func (ppu *PPU) GetColorFromPalette(palette, pixel uint8) *PPUColor {
	data := ppu.bus.Read(0x3F00 + (uint16(palette) << 2) + uint16(pixel)&0x3F)

	if data >= 0x40 {
		fmt.Printf("Trying to access palette index out of range: %d.\n", data)
		return &PPUColor{0, 0, 0}
	}

	return colorTable[data]
}

type setPixel func(x, y uint16, pixel uint8)

func (ppu *PPU) DrawPatternTable(i uint16, draw setPixel) {
	for tileY := uint16(0); tileY < 16; tileY++ {
		for tileX := uint16(0); tileX < 16; tileX++ {
			// Convert x,y to linear offset
			offset := tileY*256 + tileX*16

			for row := uint16(0); row < 8; row++ {
				tileLSB := ppu.bus.ReadDebug(i*0x1000 + offset + row)
				tileMSB := ppu.bus.ReadDebug(i*0x1000 + offset + row + 0x0008)

				for col := uint16(0); col < 8; col++ {
					pixel := (tileLSB & 0x01) + (tileMSB & 0x01)

					tileLSB >>= 1
					tileMSB >>= 1
					draw(tileX*8+(7-col), tileY*8+row, pixel)
				}
			}
		}
	}
}

var colorTable = [0x40]*PPUColor{
	{84, 84, 84},
	{0, 30, 116},
	{8, 16, 144},
	{48, 0, 136},
	{68, 0, 100},
	{92, 0, 48},
	{84, 4, 0},
	{60, 24, 0},
	{32, 42, 0},
	{8, 58, 0},
	{0, 64, 0},
	{0, 60, 0},
	{0, 50, 60},
	{0, 0, 0},
	{0, 0, 0},
	{0, 0, 0},

	{152, 150, 152},
	{8, 76, 196},
	{48, 50, 236},
	{92, 30, 228},
	{136, 20, 176},
	{160, 20, 100},
	{152, 34, 32},
	{120, 60, 0},
	{84, 90, 0},
	{40, 114, 0},
	{8, 124, 0},
	{0, 118, 40},
	{0, 102, 120},
	{0, 0, 0},
	{0, 0, 0},
	{0, 0, 0},

	{236, 238, 236},
	{76, 154, 236},
	{120, 124, 236},
	{176, 98, 236},
	{228, 84, 236},
	{236, 88, 180},
	{236, 106, 100},
	{212, 136, 32},
	{160, 170, 0},
	{116, 196, 0},
	{76, 208, 32},
	{56, 204, 108},
	{56, 180, 204},
	{60, 60, 60},
	{0, 0, 0},
	{0, 0, 0},

	{236, 238, 236},
	{168, 204, 236},
	{188, 188, 236},
	{212, 178, 236},
	{236, 174, 236},
	{236, 174, 212},
	{236, 180, 176},
	{228, 196, 144},
	{204, 210, 120},
	{180, 222, 120},
	{168, 226, 144},
	{152, 226, 180},
	{160, 214, 228},
	{160, 162, 160},
	{0, 0, 0},
	{0, 0, 0},
}
