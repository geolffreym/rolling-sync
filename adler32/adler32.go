// Adler Rolling Checksum
// Based on rsync algorithm https://rsync.samba.org/tech_report/node3.html

package adler32

import "errors"

const M = 65521

type Adler32 struct {
	Window  []byte
	count   int    // Last position
	a, b    uint16 // adler32 formula
}

func New(size int) *Adler32 {
	return &Adler32{
		Window: []byte{},
		count:  0,
		a:      1,
		b:      0,
	}
}
func (h *Adler32) Reset() {
	h.a = 0
	h.b = 0
	h.count = 0
	h.Window = h.Window[:0]
}

// Calculate initial checksum from byte slice
func (h *Adler32) Write(data []byte) {
	//https://rsync.samba.org/tech_report/node3.html
	for index, char := range data {
		h.a += uint16(char)
		h.b += uint16(len(data)-index) * uint16(char)
		h.count++
	}

	h.a %= M
	h.b %= M

}

// Calculate and return Checksum
func (h *Adler32) Sum() uint32 {
	// x =  920 =  0x398  (base 16)
	// y = 4582 = 0x11E6
	// Output = 0x11E6 << 16 + 0x398 = 0x11E60398
	return uint32(h.b)<<16 | uint32(h.a)
}

func (h *Adler32) Count() int { return h.count }

// Add byte to rolling checksum
func (h *Adler32) RollIn(input byte) {
	h.a += uint16(input)
	h.b += h.a
	// Keep stored windows bytes while get processed
	h.Window = append(h.Window, input)
	h.count++
}

// Substract byte from checksum
func (h *Adler32) RollOut() (byte, error) {

	if len(h.Window) == 0 {
		return byte(0), errors.New("Window size equal 0")
	}

	old := h.Window[0]
	h.a = h.a - uint16(old)
	h.b = h.b - (uint16(len(h.Window)) * uint16(old))
	h.Window = h.Window[1:]
	h.count--

	return old, nil
}
