package core

import "log"

type readWriteDevice interface {
	Read(addr uint16) (uint8, bool)
	Write(addr uint16, data uint8) bool

	// Just read without modifying anything.
	ReadDebug(addr uint16) (uint8, bool)
}

type Bus struct {
	devices []readWriteDevice
}

// ConnectDevice - connects device to the bus
func (bus *Bus) ConnectDevice(device readWriteDevice) {
	bus.devices = append(bus.devices, device)
}

func (bus *Bus) Read(addr uint16) uint8 {
	// Go trough all devices - correct one will pick it up.
	for _, dev := range bus.devices {
		val, handled := dev.Read(addr)

		if handled {
			return val
		}
	}

	log.Printf("Trying to read from address 0x%X from the bus but there is no device to handle it.\n", addr)
	return 0x00
}

func (bus *Bus) ReadDebug(addr uint16) uint8 {
	// Go trough all devices - correct one will pick it up.
	for _, dev := range bus.devices {
		val, handled := dev.ReadDebug(addr)

		if handled {
			return val
		}
	}

	log.Printf("DEBUG: Trying to read from address 0x%X from the bus but there is no device to handle it.\n", addr)
	return 0x00
}

func (bus *Bus) ReadDebug16(addr uint16) uint16 {
	low := uint16(bus.ReadDebug(addr))
	high := uint16(bus.ReadDebug(addr + 1))
	return (high << 8) | low
}

func (bus *Bus) Read16(addr uint16) uint16 {
	low := uint16(bus.Read(addr))
	high := uint16(bus.Read(addr + 1))
	return (high << 8) | low
}

func (bus *Bus) Write(addr uint16, val uint8) {
	// Go trough all devices - correct one will pick it up.
	for _, dev := range bus.devices {
		handled := dev.Write(addr, val)

		if handled {
			return
		}
	}

	log.Printf("Trying to write to address 0x%X on the bus but there is no device to handle it.\n", addr)
}

func (bus *Bus) Write16(addr uint16, val uint16) {
	bus.Write(addr, uint8(val&0x00FF))
	bus.Write(addr+1, uint8((val>>8)&0x00FF))
}
