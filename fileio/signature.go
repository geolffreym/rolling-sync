package fileio

import (
	"encoding/gob"
	"errors"
	"os"

	"github.com/geolffreym/rolling-sync/sync"
)

// Write signature based on signature table
// Return error if file creation fail or encode signatures fail
func WriteSignature(file string, signatures []sync.Table) error {

	if len(signatures) == 0 {
		return errors.New("no signatures to write")
	}

	//  Performed writing operations
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()
	enc := gob.NewEncoder(f)
	err = enc.Encode(signatures)
	if err != nil {
		return err
	}

	return nil
}

// Read signatures from file and decode it
// Return error if file reading fail or decode signatures fail
func ReadSignature(file string) ([]sync.Table, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	var read []sync.Table

	dataDecoder := gob.NewDecoder(f)
	err = dataDecoder.Decode(&read)

	if err != nil {
		return nil, err
	}

	return read, nil
}
