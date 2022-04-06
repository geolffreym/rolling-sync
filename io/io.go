package io

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
)

type IO struct {
	file      *os.File
	BlockSize int
}

// Process file stats
func (r *IO) Read(input string) ([][]byte, error) {
	// Open file to split
	file, err := os.Open(input)
	if err != nil {
		return nil, err
	}

	// Get file info and get total file size
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	// Calculate file chunks availables
	fileChunks := r.Chunks(fileSize)
	fmt.Printf("Total Pieces %d \n", fileChunks)

	if fileChunks <= 1 {
		return nil, errors.New("At least 2 chunks are required")
	}

	r.file = file
	blocks, err := r.Blocks()

	if err != nil {
		return nil, err
	}

	return blocks, nil
}

// Return chunks length based on file size
func (r *IO) Chunks(fileSize int64) int {
	return int(math.Ceil(float64(fileSize) / float64(r.BlockSize)))
}

// Generate checksum blocks from file chunks
// INFO: this probably could cause memory issues for big files
// INFO: keep this approach for test only
// INFO: in real use case could be improved using *bufio.Reader
func (r *IO) Blocks() ([][]byte, error) {

	if r.file == nil {
		return nil, errors.New("No file set please Read one a first")
	}

	blocks := [][]byte{}
	file := bufio.NewReader(r.file)
	defer r.file.Close()

	for {
		//Read chunks from file
		chunkBuffer := make([]byte, r.BlockSize)
		bytesRead, err := file.Read(chunkBuffer)
		// Stop if not bytes read or end to file
		if bytesRead == 0 || err == io.EOF {
			break
		}

		// Persist checksum for blocks
		blocks = append(blocks, chunkBuffer)
	}

	return blocks, nil
}

// Write sha256 signature
func (r *IO) WriteSignature(file string) error {
	hasher := sha256.New()

	if _, err := io.Copy(hasher, r.file); err != nil {
		return err
	}

	signature, err := os.Create(file)
	if err != nil {
		return err
	}

	signature.Write(hasher.Sum(nil))
	return nil

}
