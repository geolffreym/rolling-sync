/**
Copyright (c) 2022, Geolffrey Mena <gmjun2000@gmail.com>
Circular buffer interface based on https://github.com/balena-os/circbuf
**/
package sync

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"io"

	"github.com/geolffreym/rolling-sync/adler32"
	"github.com/geolffreym/rolling-sync/utils"
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

// Struct to handle weak + strong checksum operations
type Table struct {
	Weak   uint32
	Strong string
}

type Sync struct {
	blockSize int
}

// Factory function
func New(size int) Sync {
	return Sync{
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
	weak := adler32.New()
	return weak.Write(block).Sum()
}

// Fill signature from blocks using
// Weak + Strong hash table to avoid collisions.
// Hash table improve performance for mapping search using strong calc only if weak is found
func (s Sync) BuildSigTable(reader *bufio.Reader) []Table {
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
		weak := weak(block)
		strong := strong(block)
		// Keep signatures while get written
		table := Table{Weak: weak, Strong: strong}
		signatures = append(signatures, table)
	}

	return signatures
}

// Based on weak + string map searching for block position
// in indexes and return block number or error if not found
func (s Sync) Seek(idx Indexes, wk uint32, b []byte) int {
	// Check if weaksum exists in indexes table
	if subfield, found := idx[wk]; found {
		st := strong(b) // Calc strong hash until weak found
		if _, ok := subfield[st]; ok {
			return subfield[st]
		}
	}

	return -1
}

// Fill tables indexes to match block position and return indexes:
// {weak strong} = 0, {weak, strong} = 1
func (s Sync) BuildIndexes(signatures []Table) Indexes {
	indexes := make(Indexes) // Build Indexes
	// Keep signatures in memory while get processed
	for i, check := range signatures {
		indexes[check.Weak] = map[string]int{check.Strong: i}
	}

	return indexes
}

// Check if any block get removed and return the cleaned/amplified matches copy with missing blocks
func (s Sync) IntegrityCheck(sig []Table, matches map[int]Bytes) map[int]Bytes {
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

// Return new calculated range position in block diffs
func (s Sync) block(index int, literalMatches []byte) Bytes {
	return Bytes{
		Start:  (index * s.blockSize),                 // Block change start
		Offset: ((index * s.blockSize) + s.blockSize), // Block change endwhereas it could be copied-on-write to a new data structure
		Lit:    literalMatches,                        // Store literal matches
	}
}

// Calculate "delta" and return match diffs.
// Return map "Bytes" matches, each Byte keep position and literal
// diff matches for block and the map key keep the block position.
func (s Sync) Delta(sig []Table, reader *bufio.Reader) map[int]Bytes {
	// Weak checksum adler32
	weak := adler32.New()
	// Indexes for block positionAppend block to match diffing list
	indexes := s.BuildIndexes(sig)
	// Literal matches keep literal diff bytes stored
	tmpLitMatches := []byte{}
	delta := make(map[int]Bytes)

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
			delta[index] = newBlock // Add new block to delta matches

			// Clear garbage collectable
			utils.Clear(&tmpLitMatches) // Clear tmp literal matches
			utils.Clear(&weak)          // Clear weak adler object
		}

	}

	// Missing blocks?
	// Finally check the blocks integrity
	// Return cleaned/amplified copy for delta matches
	delta = s.IntegrityCheck(sig, delta)
	return delta

}
