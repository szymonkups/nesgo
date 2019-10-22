package core

import (
	"log"
)

type readWriteDevice interface {
	Read(busId string, addr uint16, debug bool) (uint8, bool)
	Write(busId string, addr uint16, data uint8, debug bool) bool
}

type bus struct {
	id      string
	devices []readWriteDevice
}

func NewCPUBus() *bus {
	return &bus{id: "cpu"}
}

func NewPPUBus() *bus {
	return &bus{id: "PPU"}
}

// ConnectDevice - connects device to the bus
func (bus *bus) ConnectDevice(device readWriteDevice) {
	bus.devices = append(bus.devices, device)
}

func (bus *bus) Read(addr uint16) uint8 {
	return bus.readFromBus(addr, false)
}

func (bus *bus) Read16(addr uint16) uint16 {
	low := uint16(bus.Read(addr))
	high := uint16(bus.Read(addr + 1))
	return (high << 8) | low
}

func (bus *bus) ReadDebug(addr uint16) uint8 {
	return bus.readFromBus(addr, true)
}

func (bus *bus) ReadDebug16(addr uint16) uint16 {
	low := uint16(bus.ReadDebug(addr))
	high := uint16(bus.ReadDebug(addr + 1))
	return (high << 8) | low
}

func (bus *bus) Write(addr uint16, val uint8) {
	bus.writeToBus(addr, val, false)
}

func (bus *bus) Write16(addr uint16, val uint16) {
	bus.Write(addr, uint8(val&0x00FF))
	bus.Write(addr+1, uint8((val>>8)&0x00FF))
}

func (bus *bus) WriteDebug(addr uint16, val uint8) {
	bus.writeToBus(addr, val, true)
}

func (bus *bus) readFromBus(addr uint16, debug bool) uint8 {
	// Go trough all devices - correct one will pick it up.
	for _, dev := range bus.devices {
		val, handled := dev.Read(bus.id, addr, debug)

		if handled {
			return val
		}
	}

	log.Printf("Trying to read from address 0x%X from the bus \"%s\" but there is no device to handle it.\n", addr, bus.id)
	return 0x00
}

func (bus *bus) writeToBus(addr uint16, val uint8, debug bool) {
	// Go trough all devices - correct one will pick it up.
	for _, dev := range bus.devices {
		handled := dev.Write(bus.id, addr, val, debug)

		if handled {
			return
		}
	}

	log.Printf("Trying to write to address 0x%X on the bus \"%s\" but there is no device to handle it.\n", addr, bus.id)
}
