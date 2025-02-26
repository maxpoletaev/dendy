package genie

var (
	letters       = []byte("APZLGITYEOXUKSVN")
	letterMapping = make(map[byte]uint16)
)

func init() {
	// A:0x1, E:0x2 ... N:0xF
	for i, l := range letters {
		letterMapping[l] = uint16(i)
	}
}

// Game Genie code decoding. Source:
// https://tuxnes.sourceforge.net/gamegenie.html

func decode6(code []byte) (uint16, override) {
	if len(code) != 6 {
		panic("invalid code length")
	}

	var (
		n0 = letterMapping[code[0]]
		n1 = letterMapping[code[1]]
		n2 = letterMapping[code[2]]
		n3 = letterMapping[code[3]]
		n4 = letterMapping[code[4]]
		n5 = letterMapping[code[5]]
	)

	addr := 0x8000 + ((n3 & 7) << 12) |
		((n5 & 7) << 8) |
		((n4 & 8) << 8) |
		((n2 & 7) << 4) |
		((n1 & 8) << 4) |
		(n4 & 7) | (n3 & 8)

	data := ((n1 & 7) << 4) |
		((n0 & 8) << 4) |
		(n0 & 7) |
		(n5 & 8)

	return addr, override{
		data: byte(data),
		mode: 6,
	}
}

func decode8(code []byte) (uint16, override) {
	if len(code) != 8 {
		panic("invalid code length")
	}

	var (
		n0 = letterMapping[code[0]]
		n1 = letterMapping[code[1]]
		n2 = letterMapping[code[2]]
		n3 = letterMapping[code[3]]
		n4 = letterMapping[code[4]]
		n5 = letterMapping[code[5]]
		n6 = letterMapping[code[6]]
		n7 = letterMapping[code[7]]
	)

	addr := 0x8000 + ((n3 & 7) << 12) |
		((n5 & 7) << 8) |
		((n4 & 8) << 8) |
		((n2 & 7) << 4) |
		((n1 & 8) << 4) |
		(n4 & 7) | (n3 & 8)

	data := ((n1 & 7) << 4) |
		((n0 & 8) << 4) |
		(n0 & 7) |
		(n7 & 8)

	cmp := ((n7 & 7) << 4) |
		((n6 & 8) << 4) |
		(n6 & 7) |
		(n5 & 8)

	return addr, override{
		data: byte(data),
		cmp:  byte(cmp),
		mode: 8,
	}
}
