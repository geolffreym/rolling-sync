// Adler Rolling Checksum
// Based on rsync algorithm https://rsync.samba.org/tech_report/node3.html

package adler32

import (
	"hash"
	"hash/adler32"
)

const M = 65521
const S = 1 << 6 //  bits

type Adler32Checksum struct {
	window  []byte // Current window
	last    int    // Last position
	a, b, n uint32 // adler32 formula
	hash    hash.Hash32
}

func New() *Adler32Checksum {
	return &Adler32Checksum{
		window: make([]byte, 0, S),
		hash:   adler32.New(),
		last:   0,
		a:      1,
		b:      0,
		n:      0,
	}
}
func (h *Adler32Checksum) Reset() {
	h.a = 1
	h.b = 0
	h.n = 0
	h.last = 0
	h.window = h.window[:0]
	h.hash.Reset()
}

// Keep  window chunk stored while get processed
func (h *Adler32Checksum) Write(data []byte) {
	h.window = data

	h.hash.Reset()
	h.hash.Write(h.window)

	s := h.hash.Sum32()
	h.a, h.b = s&0xffff, s>>16
	h.n = uint32(len(h.window)) % M
}

// Calculate and return Checksum
func (h *Adler32Checksum) Sum() uint32 {
	// a =  920 =  0x398  (base 16)
	// b = 4582 = 0x11E6
	// Output = 0x11E6 << 16 + 0x398 = 0x11E60398
	// 0xffff = 65535 = 2^16 = the largest prime number smaller than 2^16
	// Ensure you're working with 16 bits only you discard the rest by AND-ing with 0xFFFF
	// 2^16 exponential = h.b << 16
	return uint32(h.b)<<16 | uint32(h.a)
}

// Roll position a = [0123456] = (a - 0 + 7) = [1234567]
func (h *Adler32Checksum) Roll(input byte) byte {
	new := uint32(input)
	old := h.window[h.last]
	leave := uint32(old)

	// Move last pos => +1 and keep stored last input in window
	h.window[h.last] = input
	h.last++

	// Reset window position
	if h.last > len(h.window) {
		h.last = 0
	}

	h.a = (h.a + M + new - leave) % M
	h.b = (h.b + (h.n*leave/M+1)*M + h.a - (h.n * leave) - 1) % M
	return old

}
