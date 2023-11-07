package runtime

import (
	"errors"
	"go.bug.st/serial"
	"harnsgateway/pkg/runtime"
)

var ErrModbusBadConn = errors.New("Bad Modbus connection\n")
var ErrModbusServerBadResp = errors.New("Modbus server bad response\n")
var ErrModbusServerClosed = errors.New("Modbus server closed\n")
var ErrMessageTransaction = errors.New("Modbus message transaction not match\n")
var ErrMessageSlave = errors.New("Modbus message slave not match\n")
var ErrMessageDataLengthNotEnough = errors.New("Modbus message data length not enough\n")
var ErrMessageFunctionCodeError = errors.New("Modbus message function code error\n")
var ErrManyRetry = errors.New("Connect Modbus server retry more than three times\n")
var ErrDeviceType = errors.New("Error device type\n")
var ErrCRC16Error = errors.New("Validate crc16 error\n")

type ModbusModel byte

const (
	Tcp ModbusModel = iota
	Rtu
	RtuOverTcp
)

//	var ModbusModelToString = map[ModbusModel]string{
//		Tcp:        "modbusTcp",
//		Rtu:        "modbusRtu",
//		RtuOverTcp: "modbusRtuOverTcp",
//	}
var StringToModbusModel = map[string]ModbusModel{
	"modbusTcp":        Tcp,
	"modbusRtu":        Rtu,
	"modbusRtuOverTcp": RtuOverTcp,
}

type FunctionCode uint8

const (
	ReadCoilStatus FunctionCode = iota + 1
	ReadInputStatus
	ReadHoldRegister
	ReadInputRegister
	WriteSingleCoil
	WriteSingleRegister
	NOON7
	NOON8
	NOON9
	NOON10
	NOON11
	NOON12
	NOON13
	NOON14
	WriteMultipleCoil
	WriteMultipleRegister
)

const (
	// PerRequestMaxCoil functionCode01 一次最多读取248个字节 总共248 * 8 = 1984个线圈
	PerRequestMaxCoil = 1983
	// PerRequestMaxRegister functionCode03 一次最多读取124个寄存器,248个字节
	PerRequestMaxRegister = 123
)

var StopBitsToStopBits = map[runtime.StopBits]serial.StopBits{
	runtime.OneStopBit:           serial.OneStopBit,
	runtime.OnePointFiveStopBits: serial.OnePointFiveStopBits,
	runtime.TwoStopBits:          serial.TwoStopBits,
}

var ParityToParity = map[runtime.Parity]serial.Parity{
	runtime.NoParity:   serial.NoParity,
	runtime.OddParity:  serial.OddParity,
	runtime.EvenParity: serial.EvenParity,
}
