package main

import (
	"encoding/binary"
	"fmt"
	"log"

	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

// LAN9252 SPI Commands
const (
	CmdRead  byte = 0x03
	CmdWrite byte = 0x02
)

// LAN9252 Hardware Registers
const (
	RegByteTest uint16 = 0x0064 // Expected default value: 0x87654321
)

func main() {
	// 1. Initialize the periph host driver
	if _, err := host.Init(); err != nil {
		log.Fatalf("failed to initialize periph host: %v", err)
	}

	// 2. Open the SPI port (usually /dev/spidev0.0 on Raspberry Pi)
	portCloser, err := spireg.Open("SPI0.0")
	if err != nil {
		log.Fatalf("failed to open SPI port: %v", err)
	}
	defer portCloser.Close()

	// 3. Connect to the device with specific SPI settings
	// LAN9252 supports SPI Mode 0 and Mode 3 up to 30 MHz
	conn, err := portCloser.Connect(30*physic.MegaHertz, spi.Mode3, 8)
	if err != nil {
		log.Fatalf("failed to connect to SPI device: %v", err)
	}

	// 4. Read the BYTE_TEST register to verify communication
	val, err := readReg32(conn, RegByteTest)
	if err != nil {
		log.Fatalf("failed to read register: %v", err)
	}

	fmt.Printf("LAN9252 BYTE_TEST Register (0x%04X): 0x%08X\n", RegByteTest, val)
	if val == 0x87654321 {
		fmt.Println("Success! Communication with LAN9252 verified.")
	} else {
		fmt.Println("Error: Invalid register value returned. Check wiring or SPI mode.")
	}
}

// readReg32 reads a 32-bit register from the LAN9252
func readReg32(conn spi.Conn, regAddr uint16) (uint32, error) {
	// LAN9252 Read Protocol: Instruction (1 byte) + Address (2 bytes) + Dummy (1 byte) + Data (4 bytes)
	tx := make([]byte, 4+4)
	rx := make([]byte, len(tx))

	tx[0] = CmdRead
	tx[1] = byte(regAddr >> 8)
	tx[2] = byte(regAddr & 0xFF)
	tx[3] = 0x00 // Dummy byte required by LAN9252

	// Perform full-duplex SPI transaction
	if err := conn.Tx(tx, rx); err != nil {
		return 0, err
	}

	// Extract the 4 data bytes (LAN9252 uses Little-Endian format for registers)
	dataBytes := rx[4:]
	return binary.LittleEndian.Uint32(dataBytes), nil
}

// writeReg32 writes a 32-bit value to a register on the LAN9252
func writeReg32(conn spi.Conn, regAddr uint16, value uint32) error {
	// LAN9252 Write Protocol: Instruction (1 byte) + Address (2 bytes) + Data (4 bytes)
	tx := make([]byte, 3+4)
	rx := make([]byte, len(tx))

	tx[0] = CmdWrite
	tx[1] = byte(regAddr >> 8)
	tx[2] = byte(regAddr & 0xFF)
	
	binary.LittleEndian.PutUint32(tx[3:], value)

	return conn.Tx(tx, rx)
}
