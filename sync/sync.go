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
	checksums  map[uint32]map[string]int
	s          hash.Hash
	w          *adler32.Adler32
}

func New(size int) *Sync {
	return &Sync{
		blockSize: size,
		checksums: make(map[uint32]map[string]int),
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

func (s *Sync) Delta(signatures []Table, reader *bufio.Reader) (delta []byte) {
	// var n int
	var notFound error
	var bytesRead int
	var err error

	s.fill(signatures)
	block := make([]byte, s.blockSize)
	bytesRead, err = reader.Read(block)

	// read:
	for {
		s.w.Reset()
		s.w.Write(block)

		for {
			w := s.w.Sum()
			_, notFound = s.seek(w, block)

			if notFound == nil {
				fmt.Printf("%s\n", block)
				block = make([]byte, s.blockSize)
				bytesRead, err = reader.Read(block)
				break
			}

			bytesRead--
			c, e := reader.ReadByte()
			fmt.Printf("%s", e)
			if e != nil {
				break
			}

			delta = append(delta, s.w.Roll(c))

		}

		// Stop if not bytes read or end to file
		if bytesRead == 0 || err == io.EOF {
			break
		}
	}

	return

}

// Return signatures tables
func (s *Sync) Signatures() []Table {
	return s.signatures
}
