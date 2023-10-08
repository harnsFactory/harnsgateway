package main

import (
	"fmt"
	"harnsgateway/pkg/utils/binutil"
	"harnsgateway/pkg/utils/crcutil"
)

func main() {
	bytes := []byte{
		0x01,
		0x03,
		0x00,
		0x00,
		0x00,
		0x0A,
	}
	sum := crcutil.CheckCrc16sum(bytes)
	crc16 := make([]byte, 2)
	binutil.WriteUint16(crc16, sum)
	fmt.Println(crc16)
	crc16Correct := []byte{0xC5, 0xCD}
	fmt.Println(crc16Correct)

	bytes2 := []byte{
		0x01,
		0x03,
		0x14,
		0x00,
		0x90,
		0x00,
		0x00,
		0x00,
		0x00,
		0x01,
		0x4D,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x03,
		0x09,
		0x00,
		0x00,
		0x00,
		0x00,
	}
	sum2 := crcutil.CheckCrc16sum(bytes2)
	fmt.Println(sum2)
	crc16Correct2 := []byte{0x7F, 0x04}
	endian := binutil.ParseUint16BigEndian(crc16Correct2)
	fmt.Println(endian)
}
