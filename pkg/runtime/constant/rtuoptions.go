package constant

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
