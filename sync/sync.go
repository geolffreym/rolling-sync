package sync

import (
	"crypto/md5"
	"encoding/gob"
	"errors"
	"hash/adler32"
	"os"
)

type Table struct {
	Weak   uint32
	Strong []byte
}
type Sync struct {
	blockSize  int
	signatures []Table
}

func New(size int) *Sync {
	return &Sync{
		blockSize: size,
	}
}

// Fill signature from blocks
func (s *Sync) Fill(blocks [][]byte) {
	for _, block := range blocks {
		// Weak ans strong checksum
		// https://rsync.samba.org/tech_report/node3.html
		adler := adler32.New()
		md5 := md5.New()
		adler.Write(block)
		md5.Write(block)

		weak := adler.Sum32()
		strong := md5.Sum(nil)

		// Keep signatures while get written
		s.signatures = append(
			s.signatures,
			Table{Weak: weak, Strong: strong},
		)
	}

}

// Write signature
func (s *Sync) Write(file string) error {

	if len(s.signatures) == 0 {
		return errors.New("No signatures to write")
	}

	//  Performed writing operations
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()
	enc := gob.NewEncoder(f)
	err = enc.Encode(s.signatures)
	return nil
}

// Read signature
func (s *Sync) Read(file string) ([]Table, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	read := []Table{}
	dataDecoder := gob.NewDecoder(f)
	err = dataDecoder.Decode(&read)

	if err != nil {
		return nil, err
	}

	return read, nil
}

// Return signatures tables
func (s *Sync) Signatures(block []byte) []Table {
	return s.signatures
}
