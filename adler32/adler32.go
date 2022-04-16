// Adler Rolling Checksum
// Based on rsync algorithm https://rsync.samba.org/tech_report/node3.html

package adler32

// The sums are done modulo 65521 (the largest prime number smaller than 2^16).
const M = 65521

type Adler32 struct {
	Window []byte
	count  int // Last position
	old    uint8
	a, b   uint16 // adler32 formula
}

func New() Adler32 {
	return Adler32{
		Window: []byte{},
		count:  0,
		a:      0,
		b:      0,
	}
}

// Calculate initial checksum from byte slice
func (h Adler32) Write(data []byte) Adler32 {
	//https://en.wikipedia.org/wiki/Adler-32
	//https://rsync.samba.org/tech_report/node3.html
	for index, char := range data {
		h.a += uint16(char)
		h.b += uint16(len(data)-index) * uint16(char)
		h.count++
	}

	h.a %= M
	h.b %= M
	return h
}

// Calculate and return Checksum
func (h Adler32) Sum() uint32 {
	// Enforce 16 bits
	// a =  920 =  0x398  (base 16)
	// b = 4582 = 0x11E6
	// Output = 0x11E6 << 16 + 0x398 = 0x11E60398
	return uint32(h.b)<<16 | uint32(h.a)&0xFFFFF
}

func (h Adler32) Count() int { return h.count }

// Add byte to rolling checksum
func (h Adler32) RollIn(input byte) Adler32 {
	h.a = (h.a + uint16(input)) % M
	h.b = (h.b + h.a) % M
	// Keep stored windows bytes while get processed
	h.Window = append(h.Window, input)
	h.count++
	return h
}

// Substract byte from checksum
func (h Adler32) RollOut() Adler32 {

	if len(h.Window) == 0 {
		h.count = 0
		return h
	}

	h.old = h.Window[0]
	h.a = (h.a - uint16(h.old)) % M
	h.b = (h.b - (uint16(len(h.Window)) * uint16(h.old))) % M
	h.Window = h.Window[1:]
	h.count--

	return h
}
