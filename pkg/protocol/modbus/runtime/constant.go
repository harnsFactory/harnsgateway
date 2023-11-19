package runtime

import (
	"errors"
	"go.bug.st/serial"
	"harnsgateway/pkg/runtime/constant"
)

var ErrModbusBadConn = errors.New("bad Modbus connection")
var ErrModbusServerBadResp = errors.New("modbus server bad response")
var ErrMessageTransaction = errors.New("modbus message transaction not match")
var ErrMessageSlave = errors.New("modbus message slave not match")
var ErrMessageDataLengthNotEnough = errors.New("modbus message data length not enough")
var ErrMessageFunctionCodeError = errors.New("modbus message function code error")
var ErrManyRetry = errors.New("connect Modbus server retry more than three times")
var ErrCRC16Error = errors.New("validate crc16 error")

type ModbusModel byte

const (
	Tcp ModbusModel = iota
	Rtu
	RtuOverTcp
)

var ModbusModelToString = map[ModbusModel]string{
	Tcp:        "modbusTcp",
	Rtu:        "modbusRtu",
	RtuOverTcp: "modbusRtuOverTcp",
}
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

var StopBitsToStopBits = map[constant.StopBits]serial.StopBits{
	constant.OneStopBit:           serial.OneStopBit,
	constant.OnePointFiveStopBits: serial.OnePointFiveStopBits,
	constant.TwoStopBits:          serial.TwoStopBits,
}

var ParityToParity = map[constant.Parity]serial.Parity{
	constant.NoParity:   serial.NoParity,
	constant.OddParity:  serial.OddParity,
	constant.EvenParity: serial.EvenParity,
}
