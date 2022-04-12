/**
Copyright (c) 2022, Geolffrey Mena <gmjun2000@gmail.com>
Circular buffer interface based on https://github.com/balena-os/circbuf
**/
package sync

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"rolling/rolling"
)

const S = 16

type Table struct {
	Weak   uint32
	Strong string
}

type Sync struct {
	blockSize  int
	cursor     int
	written    int
	total      int
	delta      []byte
	cyclic     []byte
	signatures []Table
	s          hash.Hash
	w          rolling.Rolling
	match      []byte
	checksums  map[uint32]map[string]int
}

func New(size int) *Sync {
	return &Sync{
		cursor:    0,
		written:   0,
		total:     0,
		blockSize: size,
		match:     []byte{},
		checksums: make(map[uint32]map[string]int),
		cyclic:    make([]byte, size),
		delta:     make([]byte, size),
		s:         md5.New(),
		w:         *rolling.New(size),
	}
}

// Fill signature from blocks
// Weak + Strong hash table to avoid collisions
func (s *Sync) FillTable(reader *bufio.Reader) {

	for {
		//Read chunks from file
		block := make([]byte, s.blockSize)
		bytesRead, err := reader.Read(block)
		// Stop if not bytes read or end to file
		if bytesRead == 0 || err == io.EOF {
			break
		}

		// Weak and strong checksum
		// https://rsync.samba.org/tech_report/node3.html
		fmt.Printf("%s\n", block)
		weak := s.weak(block)
		strong := s.strong(block)
		// Keep signatures while get written
		s.signatures = append(
			s.signatures,
			Table{Weak: weak, Strong: strong},
		)

	}
}

// Calc strong md5 checksum
func (s *Sync) strong(block []byte) string {
	s.s.Write(block)
	defer s.s.Reset()
	return hex.EncodeToString(s.s.Sum(nil))
}

// Calc weak adler32 checksum
func (s *Sync) weak(block []byte) uint32 {
	s.w.Reset()
	s.w.Write(block)
	return s.w.Sum()
}

// Seek indexes in table and return block number
func (s *Sync) seek(w uint32, block []byte) (int, error) {
	if subfield, found := s.checksums[w]; found {
		st := s.strong(block)
		return subfield[st], nil
	}

	return 0, errors.New("Not index in hash table")
}

// Populate checksum tables
func (s *Sync) fill(signatures []Table) {
	// Keep signatures in memory while get processed
	s.signatures = signatures
	for i, check := range signatures {
		d := make(map[string]int)
		s.checksums[check.Weak] = d
		d[check.Strong] = i
	}
}

// Clear state
func (s *Sync) Reset() {
	s.w.Reset()    // clean checksum
	s.resetBytes() // clean local bytes cache
}

func (s *Sync) resetBytes() {
	s.cursor = 0
	s.written = 0
}

// WriteByte writes a single byte into the buffer.
func (s *Sync) writeByte(c byte) {
	s.cyclic[s.cursor] = c
	// base 0 restart count on overflow case eg.
	// 30 + 1 % 32 = 31
	// 31 + 1 % 32 = 0
	// 32 + 1 % 32 = 1
	// 33 + 1 % 32 = 2
	s.cursor = (s.cursor + 1) % s.blockSize
	s.written++
	s.total++
}

// Bytes provides a slice of the bytes written
func (s *Sync) Delta(signatures []Table, reader *bufio.Reader) []byte {
	s.fill(signatures)
	s.w.Reset()

	// TAIL:
	for {
		// Get byte from reader
		// eg. reader = [abcd], byte = a...
		c, err := reader.ReadByte()
		if err == io.EOF || err != nil {
			break
		}

		// Add new el to checksum
		s.w.RollIn(c)
		// Tmp store byte in memory
		s.writeByte(c)

		// Wait until we have a full bytes length
		if s.w.Count() < s.blockSize {
			continue
		}

		// If written bytes overflow current size and not match found
		// Start moving window over data
		if s.w.Count() > s.blockSize {
			// Subtract initial byte to switch left <<  bytes
			// eg. data=abcdef, window=4 => [abcd]: a << [bcd] << e
			removed, _ := s.w.RollOut()
			// Store literal matches
			s.match = append(s.match, removed)

		}

		// Checksum
		w := s.w.Sum()
		// Check if weak and strong match in signatures
		_, notFound := s.seek(w, s.cyclic)
		if notFound == nil {
			s.Reset()
		}

	}

	fmt.Printf("%s", s.match)
	return s.match

}

// // Bytes provides a slice of the bytes written. This
// // slice should not be written to. The underlying array
// // may point to data that will be overwritten by a subsequent
// // call to Bytes. It does no allocation.
// func (s *Sync) Bytes() []byte {
// 	switch {
// 	case s.written >= s.blockSize && s.cursor == 0:
// 		return s.cyclic
// 	case s.written > s.blockSize:
// 		copy(b.out, b.data[b.writeCursor:])
// 		copy(b.out[b.size-b.writeCursor:], b.data[:b.writeCursor])
// 		return b.out
// 	default:
// 		return b.data[:b.writeCursor]
// 	}
// }

// Return signatures tables
func (s *Sync) Signatures() []Table {
	return s.signatures
}
