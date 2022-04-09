package fileio

import (
	"encoding/gob"
	"errors"
	"os"
	"rolling/sync"
)

type Signature struct{}

// Write signature
func (s *Signature) Write(file string, signatures []sync.Table) error {

	if len(signatures) == 0 {
		return errors.New("No signatures to write")
	}

	//  Performed writing operations
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()
	enc := gob.NewEncoder(f)
	err = enc.Encode(signatures)
	return nil
}

// Read signature
func (s *Signature) Read(file string) ([]sync.Table, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	read := []sync.Table{}
	dataDecoder := gob.NewDecoder(f)
	err = dataDecoder.Decode(&read)

	if err != nil {
		return nil, err
	}

	return read, nil
}
