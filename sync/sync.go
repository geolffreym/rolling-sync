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

// Bytes store block differences
// Missing = false && len(Lit) > 0 (block exist and some changes made to block)
// Missing == false && Lit == nil (block intact just copy it)
// Missing = true (block missing)
type Bytes struct {
	Offset  int    // End of diff position in block
	Start   int    // Start of diff position in block
	Missing bool   // Block not found
	Lit     []byte // Literal bytes to replace
}

// Struct to handle weak + strong checksum operations
type Table struct {
	Weak   uint32
	Strong string
}

type Sync struct {
	blockSize  int
	signatures []Table
	s          hash.Hash       // Strong signature module
	w          adler32.Adler32 // weak signature module
	checksums  map[uint32]map[string]int
}

func New(size int) *Sync {
	return &Sync{
		blockSize: size,
		checksums: make(map[uint32]map[string]int),
		s:         sha1.New(),
		w:         *adler32.New(),
	}
}

// Fill signature from blocks
// Weak + Strong hash table to avoid collisions + perf
func (s *Sync) FillTable(reader *bufio.Reader) {
	//Read chunks from file
	block := make([]byte, s.blockSize)

	for {
		// Add chunks to buffer
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
	defer s.s.Reset() // Reset after call Sum and return encoded digest
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

// Populate checksum tables to match block position
// {weak strong} = 0, {weak, strong} = 1
func (s *Sync) fillChecksum(signatures []Table) {
	// Keep signatures in memory while get processed
	s.signatures = signatures
	for i, check := range signatures {
		d := make(map[string]int)
		s.checksums[check.Weak] = d
		d[check.Strong] = i
	}
}

// Check if any block get removed
func (s *Sync) IntegrityCheck(matches map[int]*Bytes) map[int]*Bytes {
	for i := range s.signatures {
		if _, ok := matches[i]; !ok {
			matches[i] = &Bytes{
				Missing: true,                            // Block not found
				Start:   i * s.blockSize,                 // Start range of block to copy
				Offset:  (i * s.blockSize) + s.blockSize, // End block to copy
			}
		}
	}

	return matches
}

// Calculate matches ranges bytes for differences
func (s *Sync) flushMatch(block int, match *Bytes) *Bytes {
	// Store matches
	match.Start = (block * s.blockSize)        // Block change start
	match.Offset = (match.Start + s.blockSize) // Block change endwhereas it could be copied-on-write to a new data structureAppend block to match diffing list
	return match
}

// Calculate "delta" and return match diffs
// Return map "Bytes" matches, each Byte keep position and literal diff matches
func (s *Sync) Delta(signatures []Table, reader *bufio.Reader) map[int]*Bytes {
	// Populate checksum block position based on weak + strong signatures
	s.fillChecksum(signatures)
	// Reset weak module state
	s.w.Reset()
	// New struct for diff state handling
	match := &Bytes{}
	matches := make(map[int]*Bytes)

	// Keep tracking changes
	for {
		// Get byte from reader
		// eg. reader = [abcd], byte = a...
		c, err := reader.ReadByte()
		// If reader == end of file or error trying to get byte
		if err == io.EOF || err != nil {
			break
		}

		// Add new el to checksum
		s.w.RollIn(c)
		// Keep moving forward if not data ready to analize
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
			match.Lit = append(match.Lit, removed)
		}

		// Checksum
		w := s.w.Sum()
		// Check if weak and strong match in signatures
		// Match found upgrade block
		index, notFound := s.seek(w, s.w.Window)
		if notFound == nil {
			// Process matches
			match = s.flushMatch(index, match)
			matches[index] = match // Store block matches
			// Reset state
			s.w.Reset()
			// Reset/Add new struct for new block match
			match = &Bytes{}

		}

	}

	// Finally check the blocks integrity
	// Missing blocks?
	// Return cleaned delta matches
	return s.IntegrityCheck(matches)

}

// Return signatures tables
func (s *Sync) Signatures() []Table {
	return s.signatures
}
