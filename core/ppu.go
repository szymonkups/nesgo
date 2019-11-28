package core

import (
	"fmt"
	"github.com/szymonkups/nesgo/core/ppu"
)

type PPUColor struct {
	R uint8
	G uint8
	B uint8
}

type setPixel func(x, y int16, pixel *PPUColor)

type PPU struct {
	NMI bool

	drawScreen setPixel

	scanLine        int16
	cycle           int16
	isFrameComplete bool
	bus             *bus
	ctrlRegister    *ppu.ControlRegister
	statusRegister  *ppu.StatusRegister
	maskRegister    *ppu.MaskRegister
	vRamAddress     *ppu.LoopyRegister
	tRamAddress     *ppu.LoopyRegister
	addressLatch    bool
	fineX           uint8
	dataBuffer      uint8

	bgNextTileId     uint8
	bgNextTileAttrib uint8
	bgNextTileLsb    uint8
	bgNextTileMsb    uint8

	bgShifterPatterLo uint16
	bgShifterPatterHi uint16
	bgShifterAttribLo uint16
	bgShifterAttribHi uint16
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
		vRamAddress:     new(ppu.LoopyRegister),
		tRamAddress:     new(ppu.LoopyRegister),
		fineX:           0,
		dataBuffer:      0,
	}

	newPPU.ctrlRegister.Write(0b00000000)
	newPPU.statusRegister.Write(0b00000000)
	newPPU.maskRegister.Write(0b00000000)
	newPPU.vRamAddress.Write(0)
	newPPU.tRamAddress.Write(0)
	newPPU.addressLatch = false
	newPPU.drawScreen = nil

	return newPPU
}

func (ppu *PPU) SetDrawMethod(draw setPixel) {
	ppu.drawScreen = draw
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
			ppu.addressLatch = false
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
			address := ppu.vRamAddress.Read()

			// Get is normally delayed by 1 cycle...
			toReturn := ppu.dataBuffer
			ppu.dataBuffer = ppu.bus.Read(address)

			// ...until we read from palette memory
			if address >= 0x3f00 {
				toReturn = ppu.dataBuffer
			}

			ppu.vRamAddress.Increment(ppu.ctrlRegister.GetIncrementMode())
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
			ppu.tRamAddress.SetNameTableX(ppu.ctrlRegister.GetNameTableX())
			ppu.tRamAddress.SetNameTableY(ppu.ctrlRegister.GetNameTableY())
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
			if !ppu.addressLatch {
				ppu.fineX = data & 0x07
				ppu.tRamAddress.SetCoarseX(data >> 3)
				ppu.addressLatch = true
			} else {
				ppu.tRamAddress.SetFineY(data & 0x07)
				ppu.tRamAddress.SetCoarseY(data >> 3)
				ppu.addressLatch = false
			}
			return true

		case 0x06:
			tRamValue := ppu.tRamAddress.Read()

			if !ppu.addressLatch {
				ppu.tRamAddress.Write((uint16(data&0x3F) << 8) | (tRamValue & 0x00FF))
				ppu.addressLatch = true
			} else {
				ppu.tRamAddress.Write((tRamValue & 0xFF00) | uint16(data))
				ppu.vRamAddress.Write(ppu.tRamAddress.Read())
				ppu.addressLatch = false
			}
			return true

		case 0x07:
			ppu.bus.Write(ppu.vRamAddress.Read(), data)
			ppu.vRamAddress.Increment(ppu.ctrlRegister.GetIncrementMode())
			return true
		}
	}

	return false
}

func (ppu *PPU) Clock() {
	// End VBlank
	if ppu.scanLine == -1 && ppu.cycle == 1 {
		ppu.statusRegister.SetVBlank(false)
	}

	if ppu.scanLine >= -1 && ppu.scanLine < 240 {
		if ppu.scanLine == 0 && ppu.cycle == 0 {
			ppu.cycle = 1
		}

		if ppu.scanLine == -1 && ppu.cycle == 1 {
			ppu.statusRegister.SetVBlank(false)
		}

		if (ppu.cycle >= 2 && ppu.cycle < 258) || (ppu.cycle >= 321 && ppu.cycle < 338) {
			ppu.updateShifters()

			switch (ppu.cycle - 1) % 8 {
			case 0:
				ppu.loadBgShifters()
				ppu.bgNextTileId = ppu.bus.Read(0x2000 | (ppu.vRamAddress.Read() & 0x0FFF))
			case 2:
				ppu.bgNextTileAttrib = ppu.bus.Read(
					0x23C0 |
						uint16(ppu.vRamAddress.GetNameTableY())<<11 |
						uint16(ppu.vRamAddress.GetNameTableX())<<10 |
						(uint16(ppu.vRamAddress.GetCoarseY())>>2)<<3 |
						uint16(ppu.vRamAddress.GetCoarseX())>>2)

				if ppu.vRamAddress.GetCoarseY()&0x02 != 0 {
					ppu.bgNextTileAttrib >>= 4
				}

				if ppu.vRamAddress.GetCoarseX()&0x2 != 0 {
					ppu.bgNextTileAttrib >>= 2
				}

				ppu.bgNextTileAttrib &= 0x03
			case 4:
				ppu.bgNextTileLsb = ppu.bus.Read(uint16(ppu.ctrlRegister.GetBgPatternTableAddress())<<12 + (uint16(ppu.bgNextTileId) << 4) + uint16(ppu.vRamAddress.GetFineY()))
			case 6:
				ppu.bgNextTileMsb = ppu.bus.Read(uint16(ppu.ctrlRegister.GetBgPatternTableAddress())<<12 + (uint16(ppu.bgNextTileId) << 4) + uint16(ppu.vRamAddress.GetFineY()) + 8)
			case 7:
				ppu.incrementScrollX()
			}
		}

		if ppu.cycle == 256 {
			ppu.incrementScrollY()
		}

		if ppu.cycle == 257 {
			ppu.transferAddressX()
		}

		if ppu.scanLine == -1 && ppu.cycle >= 280 && ppu.cycle < 305 {
			ppu.transferAddressY()
		}
	}

	if ppu.scanLine == 240 {
		// Post render scan line - skip it.
	}

	// Start VBlank
	if ppu.scanLine == 241 && ppu.cycle == 1 {
		ppu.statusRegister.SetVBlank(true)
		if ppu.ctrlRegister.GetNMIEnable() {
			ppu.NMI = true
		}
	}

	// Render to the screen
	if ppu.maskRegister.ShowBg {
		var bitMux uint16 = 0x8000 >> ppu.fineX
		var p0Pixel, p1Pixel uint8

		if (ppu.bgShifterPatterLo & bitMux) > 0 {
			p0Pixel = 1
		} else {
			p0Pixel = 0
		}

		if (ppu.bgShifterPatterHi & bitMux) > 0 {
			p1Pixel = 1
		} else {
			p1Pixel = 0
		}

		pixel := (p1Pixel << 1) | p0Pixel

		var pal0, pal1 uint8
		if (ppu.bgShifterAttribLo & bitMux) > 0 {
			pal0 = 1
		} else {
			pal0 = 0
		}

		if (ppu.bgShifterAttribHi & bitMux) > 0 {
			pal1 = 1
		} else {
			pal1 = 0
		}

		palette := (pal1 << 1) | pal0

		if ppu.drawScreen != nil {
			ppu.drawScreen(ppu.cycle-1, ppu.scanLine, ppu.GetColorFromPalette(palette, pixel))
		}

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

func (ppu *PPU) loadBgShifters() {
	ppu.bgShifterPatterLo = (ppu.bgShifterPatterLo & 0xFF00) | uint16(ppu.bgNextTileLsb)
	ppu.bgShifterPatterHi = (ppu.bgShifterPatterHi & 0xFF00) | uint16(ppu.bgNextTileMsb)

	if ppu.bgNextTileAttrib&0b01 > 0 {
		ppu.bgShifterAttribLo = (ppu.bgShifterAttribLo & 0xFF00) | 0xFF
	} else {
		ppu.bgShifterAttribLo = (ppu.bgShifterAttribLo & 0xFF00) | 0x00
	}

	if ppu.bgNextTileAttrib&0b10 > 0 {
		ppu.bgShifterAttribHi = (ppu.bgShifterAttribHi & 0xFF00) | 0xFF
	} else {
		ppu.bgShifterAttribHi = (ppu.bgShifterAttribHi & 0xFF00) | 0x00
	}
}

func (ppu *PPU) updateShifters() {
	if ppu.maskRegister.ShowBg {
		ppu.bgShifterPatterLo <<= 1
		ppu.bgShifterPatterHi <<= 1

		ppu.bgShifterAttribLo <<= 1
		ppu.bgShifterAttribHi <<= 1
	}
}

func (ppu *PPU) incrementScrollX() {
	if ppu.maskRegister.ShowBg || ppu.maskRegister.ShowSprites {
		coarseX := ppu.vRamAddress.GetCoarseX()

		if coarseX == 31 {
			ppu.vRamAddress.SetCoarseX(0)
			ppu.vRamAddress.SetNameTableX(^ppu.vRamAddress.GetNameTableX())
		} else {
			ppu.vRamAddress.SetCoarseX(coarseX + 1)
		}
	}
}

func (ppu *PPU) incrementScrollY() {
	if ppu.maskRegister.ShowBg || ppu.maskRegister.ShowSprites {
		fineY := ppu.vRamAddress.GetFineY()
		if fineY < 7 {
			ppu.vRamAddress.SetFineY(fineY + 1)
		} else {
			ppu.vRamAddress.SetFineY(0)

			coarseY := ppu.vRamAddress.GetCoarseY()
			if coarseY == 29 {
				ppu.vRamAddress.SetCoarseY(0)
				ppu.vRamAddress.SetNameTableY(^ppu.vRamAddress.GetNameTableY())
			} else if coarseY == 31 {
				ppu.vRamAddress.SetCoarseY(0)
			} else {
				ppu.vRamAddress.SetCoarseY(coarseY + 1)
			}
		}
	}
}

func (ppu *PPU) transferAddressX() {
	if ppu.maskRegister.ShowBg || ppu.maskRegister.ShowSprites {
		ppu.vRamAddress.SetNameTableX(ppu.tRamAddress.GetNameTableX())
		ppu.vRamAddress.SetCoarseX(ppu.tRamAddress.GetCoarseX())
	}
}

func (ppu *PPU) transferAddressY() {
	if ppu.maskRegister.ShowBg || ppu.maskRegister.ShowSprites {
		ppu.vRamAddress.SetFineY(ppu.tRamAddress.GetFineY())
		ppu.vRamAddress.SetNameTableY(ppu.tRamAddress.GetNameTableY())
		ppu.vRamAddress.SetCoarseY(ppu.tRamAddress.GetCoarseY())
	}
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

func (ppu *PPU) DrawPatternTable(i uint16, paletteId uint8, draw setPixel) {
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
					draw(int16(tileX*8+(7-col)), int16(tileY*8+row), ppu.GetColorFromPalette(paletteId, pixel))
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
