/**
Copyright (c) 2022, Geolffrey Mena <gmjun2000@gmail.com>
sync implements a circular buffer interface
based on https://github.com/balena-os/circbuf
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
	"rolling/adler32"
)

const S = 16

type Delta []interface{}
type Bytes struct {
	Offset int64
	Len    int
}

type Table struct {
	Weak   uint32
	Strong string
}

type Sync struct {
	blockSize  int
	signatures []Table
	delta      []byte
	data       []byte
	checksums  map[uint32]map[string]int
	cursor     int
	total      int
	s          hash.Hash
	w          *adler32.Adler32
}

func New(size int) *Sync {
	return &Sync{
		blockSize: size,
		checksums: make(map[uint32]map[string]int),
		data:      make([]byte, size),
		delta:     make([]byte, size),
		cursor:    0,
		total:     0,
		s:         md5.New(),
		w:         adler32.New(),
	}
}

// Fill signature from blocks
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
		// fmt.Printf("%s\n", block)
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
	for i, check := range signatures {
		d := make(map[string]int)
		s.checksums[check.Weak] = d
		d[check.Strong] = i
	}
}

func (s *Sync) resetBytes() {
	s.cursor = 0
	s.total = 0
}

// WriteByte writes a single byte into the buffer.
func (s *Sync) writeByte(c byte) {
	s.data[s.cursor] = c
	// base 0 restart count on overflow case eg.
	// 30 + 1 % 32 = 31
	// 31 + 1 % 32 = 0
	// 32 + 1 % 32 = 1
	// 33 + 1 % 32 = 2
	s.cursor = ((s.cursor + 1) % s.blockSize)
	s.total++
}

// Return bytes by index
func (s *Sync) getByIndex(i int) (byte, error) {
	switch {
	case i >= s.total || i >= s.blockSize:
		return 0, errors.New("Out of bounds index")
	case s.total > s.blockSize:
		rotateIndex := (s.cursor + i) % s.blockSize
		return s.data[rotateIndex], nil
	default:
		return s.data[i], nil
	}
}

// Bytes provides a slice of the bytes written
func (s *Sync) Bytes() []byte {
	switch {
	case s.total >= s.blockSize && s.cursor == 0:
		return s.data
	case s.total > s.blockSize:
		copy(s.delta, s.data[s.cursor:])
		copy(s.delta[s.blockSize-s.cursor:], s.data[:s.cursor])
		return s.delta
	default:
		return s.data[:s.cursor]
	}
}

func (s *Sync) Delta(signatures []Table, reader *bufio.Reader, out *bufio.Writer) error {

	var initial byte
	match := &Match{output: out}
	s.fill(signatures)
	s.w.Reset()

	// read:
	for {
		c, err := reader.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if s.total > 0 {
			// Keep first element from block
			// Use this initial byte to operate over checksum
			initial, _ = s.getByIndex(0)
		}

		// Roll checksum
		s.w.RollIn(c)
		// Tmp store byte in memory
		s.writeByte(c)
		fmt.Printf("%s\n", s.data)
		// Wait until we have a full bytes length
		if s.total < s.blockSize {
			continue
		}

		// // TODO Check if changes are made
		// // If written bytes overflow current size
		if s.total > s.blockSize {
			// Subtract initial byte to switch left <<  bytes
			// eg. [abcdefgh] = size 8 | a << [icdefgh] << i | c << [ijdefgh] << j
			// match.add(MATCH_KIND_LITERAL, uint64(initial), 1)
			// if w >
			s.w.RollOut(initial)

		}

		// Checksum
		w := s.w.Sum()
		// fmt.Printf("%d\n", w)
		// fmt.Printf("%s\n", s.data)
		// Check if weak and strong match in signatures
		_, notFound := s.seek(w, s.data)
		if notFound == nil {
			fmt.Printf("\nMatched=%s\n", s.data)
			s.w.Reset()    // clean checksum
			s.resetBytes() // clean local bytes cache
			// Stored action
			// match.add(MATCH_KIND_COPY, uint64(index*s.blockSize), uint64(s.blockSize))
		}

	}

	// Pending bytes
	for _, b := range s.Bytes() {
		fmt.Printf("%s", string(b))
		match.add(0, uint64(b), 1)
	}

	if err := match.flush(); err != nil {
		return err
	}

	out.Flush()
	return nil

}

// Return signatures tables
func (s *Sync) Signatures() []Table {
	return s.signatures
}
