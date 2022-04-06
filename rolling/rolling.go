package rolling

import (
	"fmt"

	"github.com/chmduquesne/rollinghash/rabinkarp64"
)

// Window len
const N = 4

type RollingHash struct {
	A    *rabinkarp64.RabinKarp64
	B    *rabinkarp64.RabinKarp64
	Diff map[int]byte
}

func New() *RollingHash {
	return &RollingHash{
		A:    rabinkarp64.New(),
		B:    rabinkarp64.New(),
		Diff: make(map[int]byte),
	}
}

// Compute alg to find differences
// A source, B target
func (h *RollingHash) Compute(A []byte, B []byte) {
	h.A.Reset()
	h.A.Write(A[:N])

	for i := N; i < len(A); i++ {
		// Chunk to check for differences
		window := B[i-N+1 : i+1]
		// Reset and write the window in classic
		h.B.Reset()
		h.B.Write(window)
		// Roll the incoming byte in rolling
		h.A.Roll(A[i])
		// Compare the hashes
		if h.A.Sum64() != h.B.Sum64() {
			// continue
			diff := h.Difference(A, B)
			fmt.Printf("%d", diff)
			if diff != 0 {
				if i < diff {
					i = diff
				}
			}

		}
	}
}

// Compute difference
func (h *RollingHash) Difference(a []byte, b []byte) int {
	for i, v := range a {
		if v&b[i] != v {
			return i
		}
	}

	return 0
}
