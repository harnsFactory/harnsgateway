package runtime

// ETagMaxInitialValue just a value, meaningless
const ETagMaxInitialValue int64 = 3294967296

type MemoryLayout byte

const (
	DCBA MemoryLayout = iota // little-endian
	CDAB                     // little-endian byte swap
	BADC                     // big-endian byte swap
	ABCD                     // big-endian
)

var MemoryLayoutToString = map[MemoryLayout]string{
	DCBA: "DCBA",
	CDAB: "CDAB",
	BADC: "BADC",
	ABCD: "ABCD",
}

var StringToMemoryLayout = map[string]MemoryLayout{
	"DCBA": DCBA,
	"CDAB": CDAB,
	"BADC": BADC,
	"ABCD": ABCD,
}

type DataType int8

const (
	BOOL DataType = iota
	INT16
	FLOAT32
	FLOAT64
	INT32
	INT64
	UINT16
	NUMBER
	STRING
)

var DataTypeToString = map[DataType]string{
	BOOL:    "bool",
	INT16:   "int16",
	FLOAT32: "float32",
	FLOAT64: "float64",
	INT32:   "int32",
	INT64:   "int64",
	UINT16:  "uint16",
	NUMBER:  "number",
	STRING:  "string",
}

var StringToDataType = map[string]DataType{
	"bool":    BOOL,
	"int16":   INT16,
	"float32": FLOAT32,
	"float64": FLOAT64,
	"int32":   INT32,
	"int64":   INT64,
	"uint16":  UINT16,
	"number":  NUMBER,
	"string":  STRING,
}

var DataTypeWord = map[DataType]uint{
	BOOL:    1,
	INT16:   1,
	FLOAT32: 2,
	FLOAT64: 4,
	INT32:   2,
	INT64:   4,
	UINT16:  1,
	NUMBER:  1,
	STRING:  1,
}

type StopBits int

const (
	// OneStopBit sets 1 stop bit (default)
	OneStopBit StopBits = iota
	// OnePointFiveStopBits sets 1.5 stop bits
	OnePointFiveStopBits
	// TwoStopBits sets 2 stop bits
	TwoStopBits
)

var StopBitsToString = map[StopBits]string{
	OneStopBit:           "1",
	OnePointFiveStopBits: "1.5",
	TwoStopBits:          "2",
}

var StringToStopBits = map[string]StopBits{
	"1":   OneStopBit,
	"1.5": OnePointFiveStopBits,
	"2":   TwoStopBits,
}

type Parity int

const (
	// NoParity disable parity control (default)
	NoParity Parity = iota
	// OddParity enable odd-parity check
	OddParity
	// EvenParity enable even-parity check
	EvenParity
	// MarkParity enable mark-parity (always 1) check
	MarkParity
	// SpaceParity enable space-parity (always 0) check
	SpaceParity
)

var ParityToString = map[Parity]string{
	NoParity:    "noParity",
	OddParity:   "oddParity",
	EvenParity:  "evenParity",
	MarkParity:  "markParity",
	SpaceParity: "spaceParity",
}

var StringToParity = map[string]Parity{
	"noParity":    NoParity,
	"oddParity":   OddParity,
	"evenParity":  EvenParity,
	"markParity":  MarkParity,
	"spaceParity": SpaceParity,
}
