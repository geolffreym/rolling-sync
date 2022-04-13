package fileio

import (
	"encoding/gob"
	"errors"
	"os"

	"github.com/geolffreym/rolling-sync/sync"
)

// Write signature
func WriteSignature(file string, signatures []sync.Table) error {

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
	enc.Encode(signatures)
	return nil
}

// Read signature
func ReadSignature(file string) ([]sync.Table, error) {
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
