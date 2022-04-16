package fileio

import (
	"bufio"
	"errors"
	"math"
	"os"
)

type IO struct {
	blockSize int
}

// Factory function
func New(blockSize int) IO {
	return IO{
		blockSize: blockSize,
	}
}

// Open file and ensure split to at least two chunks on any sufficiently sized data
func (o IO) Open(input string) (*bufio.Reader, error) {
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

	if fileChunks <= 1 {
		return nil, errors.New("At least 2 chunks are required")
	}

	return bufio.NewReader(file), nil

}

// Return chunks length based on file size
func (o IO) Chunks(fileSize int64) int {
	return int(math.Ceil(float64(fileSize) / float64(o.blockSize)))
}
