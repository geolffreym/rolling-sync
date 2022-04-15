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
	Lit     []byte // Literal bytes to replace in delta
}

// Struct to handle weak + strong checksum operations
type Table struct {
	Weak   uint32
	Strong string
}

type Sync struct {
	blockSize int
}

func New(size int) *Sync {
	return &Sync{
		blockSize: size,
	}
}

// Fill signature from blocks
// Weak + Strong hash table to avoid collisions + perf
func (s *Sync) FillTable(reader *bufio.Reader) []Table {
	//Read chunks from file
	block := make([]byte, s.blockSize)
	signatures := []Table{}

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
		signatures = append(
			signatures,
			Table{Weak: weak, Strong: strong},
		)
	}

	return signatures
}

// Calc strong md5 checksum
func (s *Sync) strong(block []byte) string {
	strong := sha1.New()
	strong.Write(block)
	return hex.EncodeToString(strong.Sum(nil))
}

// Calc weak adler32 checksum
func (s *Sync) weak(block []byte) uint32 {
	weak := adler32.New()
	weak.Write(block)
	return weak.Sum()
}

// Seek block in indexes and return block number or error if not found
func (s *Sync) seek(indexes map[uint32]map[string]int, weak uint32, block []byte) (int, error) {
	if subfield, found := indexes[weak]; found {
		st := s.strong(block)
		if _, ok := subfield[st]; ok {
			return subfield[st], nil
		}
	}

	return 0, errors.New("Not index in hash table")
}

// Populate tables indexes to match block position
// {weak strong} = 0, {weak, strong} = 1
func (s *Sync) indexTable(signatures []Table) map[uint32]map[string]int {
	indexes := make(map[uint32]map[string]int)
	// Keep signatures in memory while get processed
	for i, check := range signatures {
		d := make(map[string]int)
		indexes[check.Weak] = d
		d[check.Strong] = i
	}

	return indexes
}

// Check if any block get removed and return the cleaned/amplified matches with missing blocks
func (s *Sync) IntegrityCheck(signatures []Table, matches map[int]*Bytes) map[int]*Bytes {
	for i := range signatures {
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

// Calculate "delta" and return match diffs
// Return map "Bytes" matches, each Byte keep position and literal diff matches
func (s *Sync) Delta(signatures []Table, reader *bufio.Reader) map[int]*Bytes {
	// Weak checksum adler32
	weak := adler32.New()
	// Populate indexes by block position based on weak + strong signatures
	indexes := s.indexTable(signatures)
	// Literal matches keep literal diff bytes stored
	literalMatches := []byte{}
	matches := make(map[int]*Bytes)

	// Keep tracking changes
	for {
		// Get byte from reader
		// eg. reader = [abcd], byte = a...
		c, err := reader.ReadByte()
		// If reader == EOF end of file or error trying to get byte
		if err == io.EOF || err != nil {
			break
		}

		// Add new el to checksum
		weak.RollIn(c)
		// Keep moving forward if not data ready
		if weak.Count() < s.blockSize {
			continue
		}

		// Start moving window over data
		// If written bytes overflow current size and not match found
		if weak.Count() > s.blockSize {
			// Subtract initial byte to switch left <<  bytes
			// eg. data=abcdef, window=4 => [abcd]: a << [bcd] << e
			removed, _ := weak.RollOut()
			// Store literal matches
			literalMatches = append(literalMatches, removed)
		}

		checksum := weak.Sum() // Calc checksum based on rolling hash
		// Check if weak and strong match in checksums position based signatures
		index, notFound := s.seek(indexes, checksum, weak.Window)
		if notFound == nil {
			// Store block matches
			matches[index] = &Bytes{
				Start:  (index * s.blockSize),                 // Block change start
				Offset: ((index * s.blockSize) + s.blockSize), // Block change endwhereas it could be copied-on-write to a new data structureAppend block to match diffing list
				Lit:    literalMatches,
			}

			// Reset state
			literalMatches = nil
			weak.Reset()
		}

	}

	// Finally check the blocks integrity
	// Missing blocks?
	// Return cleaned/amplified delta matches
	matches = s.IntegrityCheck(signatures, matches)
	return matches

}
