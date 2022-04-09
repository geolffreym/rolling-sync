// Adler Rolling Checksum
// Based on rsync algorithm https://rsync.samba.org/tech_report/node3.html

package adler32

import (
	"hash"
	"hash/adler32"
)

const M = 65521
const S = 1 << 6 //  bits

type Adler32 struct {
	window  []byte // Current window
	last    int    // Last position
	a, b, n uint32 // adler32 formula
	hash    hash.Hash32
}

func New() *Adler32 {
	return &Adler32{
		window: make([]byte, 0, S),
		hash:   adler32.New(),
		last:   0,
		a:      1,
		b:      0,
		n:      0,
	}
}
func (h *Adler32) Reset() {
	h.a = 1
	h.b = 0
	h.n = 0
	h.last = 0
	h.window = h.window[:0]
	h.hash.Reset()
}

// Keep  window chunk stored while get processed
func (h *Adler32) Write(data []byte) int {
	// https://en.wikipedia.org/wiki/Adler-32
	// 0xffff = 65535 = 2^16 = the largest prime number smaller than 2^16
	// At any position p in the input, the state of the rolling hash will depend only on the last s bytes of the file
	h.window = data
	h.hash.Reset()
	h.hash.Write(h.window)

	s := h.hash.Sum32()
	h.a, h.b = s&0xffff, s>>16
	h.n = uint32(len(h.window)) % M
	return len(data)
}

// Calculate and return Checksum
func (h *Adler32) Sum() uint32 {
	// x =  920 =  0x398  (base 16)
	// y = 4582 = 0x11E6
	// Output = 0x11E6 << 16 + 0x398 = 0x11E60398
	return h.b<<16 | h.a
}

// Roll position a = [0123456] = (a - 0 + 7) = [1234567]
func (h *Adler32) Roll(input byte) byte {
	new := uint32(input)
	old := h.window[h.last]
	leave := uint32(old)

	// Move last pos => +1 and keep stored last input in window
	h.window[h.last] = input
	h.last++

	// https://en.wikipedia.org/wiki/Adler-32
	h.a = (h.a + M + new - leave) % M //
	h.b = (h.b + (1+(h.n*leave/M))*M + h.a - (h.n * leave) - 1) % M
	return old

}

func (h *Adler32) Last() int { return h.last }

// Roll position a = [0123456] = (a - 0 + 7) = [1234567]
func (h *Adler32) RollIn(input byte) {
	new := uint32(input)
	h.last += 1
	// https://en.wikipedia.org/wiki/Adler-32
	h.a += new
	h.b += h.a
}

// Roll position a = [0123456] = (a - 0 + 7) = [1234567]
func (h *Adler32) RollOut(input byte) {
	out := uint32(input)
	// https://en.wikipedia.org/wiki/Adler-32
	h.a -= out
	h.b = uint32(h.last) * out
	h.last -= 1
}
