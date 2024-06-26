// Copyright (c) 2022, Geolffrey Mena <gmjun2000@gmail.com>
// Circular buffer interface based on https://github.com/balena-os/circbuf
// Package sync implement a small library to match differences between files
package sync

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"io"
)

const S = 16

// Alias for nested map
type Indexes map[uint32]map[string]int

/*
Bytes store block differences

	Missing = true && len(Lit) == 0 (block missing)
	Missing = false && len(Lit) > 0 (block exist and some changes made to block)
	Missing == false && Lit == nil (block intact just copy it)
	Literal matches = any textual/literal value match found eg. "abcdef"
	No literal matches = any match found by position range in block to copy eg. Block missing && Start > 0 && Offset > 0
*/
type Bytes struct {
	Offset  int    // End of diff position in block
	Start   int    // Start of diff position in block
	Missing bool   // true if Block not found
	Lit     []byte // Literal bytes to replace in delta
}

// Store delta matches
type Delta map[int]Bytes

// Add new match to delta table
func (d Delta) Add(index int, b Bytes) {
	d[index] = b
}

// Struct to handle weak + strong checksum operations
type Table struct {
	Weak   uint32
	Strong string
}

type Sync struct {
	blockSize int
}

// Factory function
func New(size int) *Sync {
	return &Sync{
		blockSize: size,
	}
}

// Calc and return strong md5 checksum
func strong(block []byte) string {
	strong := sha1.New()
	strong.Write(block)
	return hex.EncodeToString(strong.Sum(nil))
}

// Calc and return weak adler32 checksum
func weak(block []byte) uint32 {
	weak := NewAdler32()
	return weak.Write(block).Sum()
}

// Return new calculated range position in block diffs
func (s *Sync) block(index int, literalMatches []byte) Bytes {
	return Bytes{
		Start:  (index * s.blockSize),                 // Block change start
		Offset: ((index * s.blockSize) + s.blockSize), // Block change endwhereas it could be copied-on-write to a new data structure
		Lit:    literalMatches,                        // Store literal matches
	}
}

// Fill signature from blocks using
// Weak + Strong hash table to avoid collisions.
// Hash table improve performance for mapping search using strong calc only if weak is found
func (s *Sync) BuildSigTable(reader *bufio.Reader) []Table {
	// Read chunks from file
	block := make([]byte, s.blockSize)
	// Declares Table nil slice
	var signatures []Table

	for {
		// Add chunks to buffer
		bytesRead, err := reader.Read(block)
		// Stop if not bytes read or end to file
		if bytesRead == 0 || err == io.EOF {
			break
		}

		// Weak and strong checksum
		// https://rsync.samba.org/tech_report/node3.
		weak := weak(block)
		strong := strong(block)
		// Keep signatures while get written
		table := Table{Weak: weak, Strong: strong}
		signatures = append(signatures, table)
	}

	return signatures
}

// Fill tables indexes to match block position and return indexes:
// {weak strong} = 0, {weak, strong} = 1
func (*Sync) BuildIndexes(signatures []Table) Indexes {
	indexes := make(Indexes) // Build Indexes
	// Keep signatures in memory while get processed
	for i, check := range signatures {
		indexes[check.Weak] = map[string]int{check.Strong: i}
	}

	return indexes
}

// Based on weak + string map searching for block position
// in indexes and return block number or -1 if not found.
func (*Sync) Seek(idx Indexes, wk uint32, b []byte) int {
	// Check if weaksum exists in indexes table
	if subfield, found := idx[wk]; found {
		st := strong(b) // Calc strong hash until weak found
		if _, ok := subfield[st]; ok {
			return subfield[st]
		}
	}

	return -1
}

// Check if any block get removed and return the cleaned/amplified matches copy with missing blocks
func (s *Sync) IntegrityCheck(sig []Table, matches Delta) Delta {
	for i := range sig {
		if _, ok := matches[i]; !ok {
			matches[i] = Bytes{
				Missing: true,                            // Block not found
				Start:   i * s.blockSize,                 // Start range of block to copy
				Offset:  (i * s.blockSize) + s.blockSize, // End block to copy
			}
		}
	}

	return matches
}

// Calculate "delta" and return match diffs.
// Return map "Bytes" matches, each Byte keep position and literal
// diff matches for block and the map key keep the block position.
func (s *Sync) Delta(sig []Table, reader *bufio.Reader) Delta {
	// Weak checksum adler32
	weak := NewAdler32()
	// Delta matches
	delta := make(Delta)
	// Indexes for block position
	indexes := s.BuildIndexes(sig)
	// Literal matches keep literal diff bytes stored
	var tmpLitMatches []byte

	// Keep tracking changes
	for {
		// Get byte from reader
		// eg. reader = [abcd], byte = a...
		c, err := reader.ReadByte()
		// If reach end of file or error trying to get byte
		if err == io.EOF || err != nil {
			break
		}

		// Add new el to checksum
		weak = weak.RollIn(c)
		// Keep moving forward if not data ready
		if weak.Count() < s.blockSize {
			continue
		}

		// Start moving window over data
		// If written bytes overflow current size and not match found
		if weak.Count() > s.blockSize {
			// Subtract initial byte to switch left <<  bytes
			// eg. data=abcdef, window=4 => [abcd]: a << [bcd] << e
			weak = weak.RollOut()
			removed := weak.Removed()
			// Store literal matches
			tmpLitMatches = append(tmpLitMatches, removed)
		}

		// Calc checksum based on rolling hash
		// Check if weak and strong match in checksums position based signatures
		index := s.Seek(indexes, weak.Sum(), weak.Window())
		if ^index != 0 { // match found
			// Generate new block with calculated range positions for diffing
			newBlock := s.block(index, tmpLitMatches)
			delta.Add(index, newBlock) // Add new block to delta matches
			// Clear garbage collectable
			tmpLitMatches = tmpLitMatches[:0] // clear tmp literal matches
			weak = NewAdler32()               // replace weak adler object
		}

	}

	// Missing blocks?
	// Finally check the blocks integrity
	// Return cleaned/amplified copy for delta matches
	return s.IntegrityCheck(sig, delta)

}
