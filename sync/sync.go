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
	delta      []Match
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
		delta:     make([]Match, size),
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

func (s *Sync) get(i int) byte {
	switch {
	case s.total > s.blockSize:
		return s.data[(s.cursor+i)%s.blockSize]
	default:
		return s.data[i]
	}
}

func (s *Sync) Delta(signatures []Table, reader *bufio.Reader) error {

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

		s.w.RollIn(c)
		s.writeByte(c)

		w := s.w.Sum()
		fmt.Printf("\n%d", w)
		fmt.Printf("\n%s", s.data)

		// Wait until we have a full bytes length
		if s.w.Last() < s.blockSize {
			continue
		}

		// If written bytes overflow current size
		if s.w.Last() > s.blockSize {
			fmt.Printf("%s", s.blockSize)
		}

		_, notFound := s.seek(w, s.data)
		if notFound == nil {
			fmt.Printf("Literal match: %s", s.data)
			s.w.Reset()
			s.resetBytes()
		}

	}

	return nil

}

// Return signatures tables
func (s *Sync) Signatures() []Table {
	return s.signatures
}
