/**
Copyright (c) 2022, Geolffrey Mena <gmjun2000@gmail.com>
Circular buffer interface based on https://github.com/balena-os/circbuf
**/
package sync

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"hash"
	"io"

	"github.com/geolffreym/rolling-sync/adler32"
)

const S = 16

type Bytes struct {
	Offset int
	Start  int
	Miss   bool
	Lit    []byte
}

type Table struct {
	Weak   uint32
	Strong string
}

type Sync struct {
	blockSize  int
	delta      []byte
	signatures []Table
	s          hash.Hash
	w          adler32.Adler32
	match      Bytes
	matches    map[int]Bytes
	checksums  map[uint32]map[string]int
}

func New(size int) *Sync {
	return &Sync{
		blockSize: size,
		match:     Bytes{},
		matches:   make(map[int]Bytes),
		checksums: make(map[uint32]map[string]int),
		delta:     make([]byte, size),
		s:         sha1.New(),
		w:         *adler32.New(size),
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
		// https://rsync.samba.org/tech_report/node3.
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
		if _, ok := subfield[st]; ok {
			return subfield[st], nil
		}
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

func (s *Sync) IntegrityCheck() {
	for i := range s.signatures {
		if _, ok := s.matches[i]; !ok {
			s.matches[i] = Bytes{
				Miss:   true,                            // Block not found
				Start:  i * s.blockSize,                 // Start range of block to copy
				Offset: (i * s.blockSize) + s.blockSize, // End block to copy
			}
		}
	}
}

// Process matches for bytes processed
func (s *Sync) flushMatch(block int) {
	// s.match.offset = (s.match.start + s.blockSize)
	// Store matches
	s.match.Start = (block * s.blockSize)
	s.match.Offset = (s.match.Start + s.blockSize)
	s.matches[block] = s.match
	s.match = Bytes{}

}

// Bytes provides a slice of the bytes written
func (s *Sync) Delta(signatures []Table, reader *bufio.Reader) map[int]Bytes {
	s.fill(signatures)
	s.w.Reset()
	// Keep tracking changes

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
		if s.w.Count() < s.blockSize {
			continue
		}

		// Start moving window over data
		// If written bytes overflow current size and not match found
		if s.w.Count() > s.blockSize {
			// Subtract initial byte to switch left <<  bytes
			// eg. data=abcdef, window=4 => [abcd]: a << [bcd] << e
			removed, _ := s.w.RollOut()
			// Store literal matches
			s.match.Lit = append(s.match.Lit, removed)
		}

		// Checksum
		w := s.w.Sum()
		// Check if weak and strong match in signatures
		// Match found upgrade block
		index, notFound := s.seek(w, s.w.Window)
		if notFound == nil {
			// Process matches
			s.flushMatch(index)
			// store block processed
			s.w.Reset()

		}

	}

	s.IntegrityCheck()
	return s.matches

}

// Return signatures tables
func (s *Sync) Signatures() []Table {
	return s.signatures
}
