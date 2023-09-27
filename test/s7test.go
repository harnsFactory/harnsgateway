package main

import (
	"fmt"
	"harnsgateway/pkg/utils/binutil"
	"io"
	"net"
)

func main() {
	tunnel, _ := net.Dial("tcp", "127.0.0.1:103")
	fmt.Println(tunnel.LocalAddr())
	// cotp报文
	// bytes := make([]byte, 0)
	/**
	First group
	contains device IDs for which resources are provided in the S7:
	01: PG or PC
	02: OS (operating or monitoring device)
	03: Others, such as OPC server, Simatic S7 PLC...

	Second group
	contains the addresses of these components
	Left character (bits 7....4):
	Rack number multiplied by 2
	Right character (bits 3...0):
	CPU slot (< 16). S7-300 always uses slot 2
	The standard TSAPs MUST be used on the PLC side (Dest TSAP field of the device).
	The local TSAP of the device (Own TSAP field) may be selected freely, but should have the same format. We recommend to use 01 01 in the Own TSAP field.
	Examples:
	03 02 Communication with the S7 CPU in rack 0, slot 2
	03 43 Communication with the S7 CPU in rack 2, slot 3
	03 2E Communication with the S7 CPU in rack 1, slot 14
	S7-1200
	The S7-1200 is usually addressed with the TSAP 02 01 (binary).
	S7-300
	The S7-300 is usually addressed with the TSAP 03 02 (binary).
	S7 1500
	The S7 1500 does not use the service and rack addressing.
	The own TSAP should be binary 06 00, the destination TSAP is ASCII "SIMATIC-ROOT-ES"
	*/
	bytes := []byte{
		// TPKT
		0x03,
		0x00,
		0x00,
		0x16, // 总字节数 固定22
		// COTP
		0x11,       // 当前字节以后的字节数
		0xe0,       // 连接请求
		0x00, 0x00, // DST reference
		0x00, 0x01, // SRC reference
		0x00, // extended formats
		0xc0, // parameter-code tpdu-size
		0x01, // parameter-Len 1
		0x0a, // TPDU size //1024
		0xc1, // parameter-code src-tsap
		0x02, // parameter-Len
		0x01, // Source TSAP  随意设置
		0x02, //
		0xc2, // parameter code:dst-size
		0x02, // parameter length
		0x01, // Destination TSAP connectionType   01PG  02OP 03s7单边 0x10s7双边
		0x01, // rack & solt
	}
	//
	w, _ := tunnel.Write(bytes)
	fmt.Println(w)
	response := make([]byte, 22)
	_, err2 := io.ReadAtLeast(tunnel, response, 22)
	if err2 != nil {
		fmt.Println(err2)
	}
	i := int(response[5]) // 208
	fmt.Println(i)
	// s7
	setupBytes := []byte{
		// TPKT
		0x03, 0x00,
		0x00, 0x19, // 总字节数
		// cotp
		0x02, // 当前字节以后的字节数
		0xf0, // 0xf0建立通信  0x04读取值  0x05写入值 0x29关闭PLC
		0x80, // 固定
		// s7 header
		0x32,       // 协议号
		0x01,       // 0x01主站发送请求 0x02从站响应请求 0x03从站响应请求并携带数据 0x07原始协议扩展
		0x00, 0x00, // 固定
		0x04, 0x00, // protocol data unit refrence 1024
		0x00, 0x08, // parameter length参数长度
		0x00, 0x00, // 数据长度 0
		// s7 parameter
		0xf0,       // 建立通信
		0x00,       // 固定
		0x00, 0x01, // Max AmQ calling
		0x00, 0x01, // Max AmQ calling
		0x01, 0xe0, // pdu length 480
	}
	write, err := tunnel.Write(setupBytes)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(write)
	response2 := make([]byte, 27)
	_, err2 = io.ReadAtLeast(tunnel, response2, 27)
	if err2 != nil {
		fmt.Println(err2)
	}

	j := int(response2[8]) // 3
	errClass := int(response2[17])
	errCode := int(response2[18])
	fmt.Println(j)
	fmt.Println(errClass)
	fmt.Println(errCode)

	responsePduLength := response2[25:]
	fmt.Println(binutil.ParseUint16(responsePduLength))
	// 读取数据
	dataBytes := []byte{
		// TPKT
		0x03, 0x00,
		0x00, 0x2b, // 总字节数
		// cotp
		0x02, // parameter length
		0xf0, // 设置通信
		0x80, // TPDU number
		// s7 header
		0x32, // 协议
		0x01, // 主站发送请求
		0x00, 0x00,
		0x00, 0x01,
		0x00, 0x1a, // parameter length
		0x00, 0x00, // Data length
		// s7 parameter
		0x04, // read value
		0x02, // item count
		// item[1]
		0x12,       // 结构标识
		0x0a,       // 此字节往后的字节长度
		0x10,       // Syntax Id Address data s7-any pointer
		0x04,       // Transport size 0x01 BIT 0x02 Byte 0x03 CHAR 0x04 WORD 0x05 INT 0x06 DWORD 0x07 DINT 0x08 REAL 0x09 DATE
		0x00, 0x02, // 数据长度
		0x00, 0x02, // 数据块编号 DB2  DB2.DBD24
		0x84,             // Area 0x81 I   0x82 Q  0x83 M  0x84 (DB) V  0x85 DI  0x86 L 0x87 V  0x1c C   0x1d T   0x1e IEC计数器   0x1f IEC定时器
		0x00, 0x00, 0xc0, // Byte Address(18-3) BitAdress(2-0)

		0x12,       // 结构标识
		0x0a,       // 此字节往后的字节长度
		0x10,       // Syntax Id Address data s7-any pointer
		0x04,       // Transport size 0x01 BIT 0x02 Byte 0x03 CHAR 0x04 WORD 0x05 INT 0x06 DWORD 0x07 DINT 0x08 REAL 0x09 DATE
		0x00, 0x02, // 数据长度
		0x00, 0x07, // 数据块编号 DB7  DB7.DBD30
		0x84,             // Area 0x81 I   0x82 Q  0x83 M  0x84 (DB) V  0x85 DI  0x86 L 0x87 V  0x1c C   0x1d T   0x1e IEC计数器   0x1f IEC定时器
		0x00, 0x00, 0xf0, // Byte Address(18-3) BitAdress(2-0)
	}
	dw, err := tunnel.Write(dataBytes)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(dw)
	response3 := make([]byte, 480)
	_, err2 = io.ReadAtLeast(tunnel, response3, 19)
	if err2 != nil {
		fmt.Println(err2)
	}

	dataErrClass := response3[17]
	dataErrCode := response3[18]
	fmt.Println(int(dataErrClass)) // 133 error
	fmt.Println(int(dataErrCode))

	data := response3[21:]
	item1Suceess := data[0]
	fmt.Println(item1Suceess)
	item1TransportSize := data[1]
	fmt.Println(item1TransportSize)
	item1Length := data[2:4]
	fmt.Println(binutil.ParseUint16(item1Length))
	item1Data := data[4:8]

	parseFloat32 := binutil.ParseFloat32(item1Data)
	fmt.Println(parseFloat32)
	// tunnel.Close()
	// fmt.Println(tunnel)
	// 分割线

	// 读取数据
	// dataBytes := []byte{
	// 	// TPKT                                    7
	// 	0x03, 0x00,
	// 	0x00, 0x1f, // 总字节数
	// 	// cotp
	// 	0x02, // parameter length
	// 	0xf0, // 设置通信
	// 	0x80, // TPDU number
	// 	// s7 header                                10
	// 	0x32, // 协议
	// 	0x01, // 主站发送请求
	// 	0x00, 0x00,
	// 	0x00, 0x01,
	// 	0x00, 0x0e, // parameter length
	// 	0x00, 0x00, // Data length
	// 	// s7 parameter                              2
	// 	0x04, // read value
	// 	0x01, // item count
	// 	// item[1]                                    12
	// 	0x12,       // 结构标识
	// 	0x0a,       // 此字节往后的字节长度
	// 	0x10,       // Syntax Id Address data s7-any pointer
	// 	0x02,       // Transport size 0x01 BIT 0x02 Byte 0x03 CHAR 0x04 WORD 0x05 INT 0x06 DWORD 0x07 DINT 0x08 REAL 0x09 DATE
	// 	0x00, 0x01, // 数据长度
	// 	0x00, 0x00, // 数据块编号 DB2  DB2.DBD24
	// 	0x81,             // Area 0x81 I   0x82 Q  0x83 M  0x84 (DB) V  0x85 DI  0x86 L 0x87 V  0x1c C   0x1d T   0x1e IEC计数器   0x1f IEC定时器
	// 	0x00, 0x00, 0x50, // Byte Address(18-3) BitAdress(2-0)
	// }
	// // startAddress 10 bitAddress 0 2 3 5
	// // 00000000 00000000 00000000
	//
	// // 00000000 00000000 01010000
	//
	// //
	//
	// dw, err := tunnel.Write(dataBytes)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(dw)
	// response3 := make([]byte, 480)
	// _, err2 = io.ReadAtLeast(tunnel, response3, 19)
	// if err2 != nil {
	// 	fmt.Println(err2)
	// }
	//
	// dataErrClass := response3[17]
	// dataErrCode := response3[18]
	// fmt.Println(int(dataErrClass)) // 133 error
	// fmt.Println(int(dataErrCode))
	//
	// data := response3[21:]
	// item1Suceess := data[0]
	// fmt.Println(item1Suceess)
	// item1TransportSize := data[1]
	// fmt.Println(item1TransportSize)
	// item1Length := data[2:4]
	// fmt.Println(binutil.ParseUint16(item1Length))
	// item1Data := data[4:8]
	//
	// parseFloat32 := binutil.ParseFloat32(item1Data)
	// fmt.Println(parseFloat32)
	tunnel.Close()
}
