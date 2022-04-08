package adler32

import (
	"hash"
	"hash/adler32"
)

const M = 65521
const S = 1 << 6 //  bits
type Adler32 struct {
	window     []byte
	last       int
	x, y, z, c uint32
	hash       hash.Hash32
}

func New() *Adler32 {
	return &Adler32{
		window: make([]byte, 0, S),
		hash:   adler32.New(),
		last:   0,
		x:      1,
		y:      0,
		z:      0,
	}
}
func (h *Adler32) Reset() {
	h.x = 1
	h.y = 0
	h.z = 0
	h.last = 0
	h.window = h.window[:0]
	h.hash.Reset()
}

/** Keep each window chunk stored
	window[0][0 0 0 0 0 0 0 0 ]
	window[1][0 1 2 0 0 5 0 3 ]
**/
func (h *Adler32) Write(data []byte) int {
	h.window = data
	h.hash.Reset()
	h.hash.Write(h.window)

	// https://en.wikipedia.org/wiki/Adler-32
	// 0xffff = 65535 = 2^16 = the largest prime number smaller than 2^16
	// At any position p in the input, the state of the rolling hash will depend only on the last s bytes of the file
	s := h.hash.Sum32()
	h.hash.Reset()
	h.hash.Write(h.window)
	h.x, h.y = s&0xffff, s>>16
	h.z = uint32(len(h.window)) % M
	return len(data)
}

func (h *Adler32) Sum() uint32 {
	// x =  920 =  0x398  (base 16)
	// y = 4582 = 0x11E6
	// Output = 0x11E6 << 16 + 0x398 = 0x11E60398
	return h.y<<16 | h.x
}

func (h *Adler32) Roll(input byte) byte {
	new := uint32(input)
	old := h.window[h.last]
	leave := uint32(old)

	// Move last pos => +1 and keep stored last input in window
	h.window[h.last] = input
	h.last++

	// https://en.wikipedia.org/wiki/Adler-32
	h.x = (h.x + M + new - leave) % M //
	h.y = (h.y + (h.z*leave/M+1)*M + h.x - (h.z * leave) - 1) % M
	return old

}
