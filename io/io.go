package io

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
)

type IO struct {
	blockSize int
	Signature *Signature
}

func New(blockSize int) *IO {
	return &IO{
		blockSize: blockSize,
		Signature: &Signature{},
	}
}

// Process file stats
func (o *IO) Open(input string) (*bufio.Reader, error) {
	// Open file to split
	file, err := os.Open(input)
	if err != nil {
		return nil, err
	}

	// Get file info and get total file size
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	// Calculate file chunks availables
	fileChunks := o.Chunks(fileSize)
	fmt.Printf("Total Pieces %d \n", fileChunks)

	if fileChunks <= 1 {
		return nil, errors.New("At least 2 chunks are required")
	}

	return bufio.NewReader(file), nil

}

// Return chunks length based on file size
func (o *IO) Chunks(fileSize int64) int {
	return int(math.Ceil(float64(fileSize) / float64(o.blockSize)))
}

// Generate checksum blocks from file chunks
// INFO: this probably could cause memory issues for big files
// INFO: keep this approach for test only
// INFO: in real use case could be improved using *bufio.Reader
// INFO: could be improved using concurrency file
func (o *IO) Blocks(file string) ([][]byte, error) {

	reader, err := o.Open(file)
	if err != nil {
		return nil, err
	}

	blocks := [][]byte{}

	for {
		//Read chunks from file
		chunkBuffer := make([]byte, o.blockSize)
		bytesRead, err := reader.Read(chunkBuffer)
		// Stop if not bytes read or end to file
		if bytesRead == 0 || err == io.EOF {
			break
		}

		// Persist checksum for blocks
		blocks = append(blocks, chunkBuffer)
	}

	return blocks, nil
}
