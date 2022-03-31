package io

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"os"
)

type IO struct {
	FileDir       string
	BufferSize    int
	FileSize      int64
	FileChunks    int
	SignatureFile string
	file          *os.File
}

func (r *IO) Read() (*IO, error) {
	// Open file to split
	file, err := os.Open(r.FileDir)
	if err != nil {
		return r, err
	}

	// Close after routine finish
	defer file.Close()

	// Get file info and get total file size
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	// Calculate file chunks availables
	fileChunks := int(math.Ceil(float64(fileSize) / float64(r.BufferSize)))
	fmt.Printf("Total Pieces %d \n", fileChunks)

	if fileChunks <= 1 {
		return r, errors.New("At least 2 chunks are required")
	}

	r.file = file
	r.FileChunks = fileChunks
	r.FileSize = fileSize
	return r, nil
}

// Receive as input a file and return hashed chunks
func (r *IO) WriteSignature() error {

	if r.FileChunks == 0 {
		return errors.New("No files chunk to process")
	}

	f, err := os.Create(r.SignatureFile)
	if err != nil {
		return err
	}

	// Iterate over chunks and hash(chunk)
	for cursor := 0; cursor < r.FileChunks; cursor++ {
		// Read chunks from file
		// Keep cursor moving from chunks in file and get smaller chunk if not % 2
		chunkSize := math.Ceil(math.Min(float64(r.BufferSize), float64(r.FileSize-int64(r.BufferSize*cursor))))
		chunkBuffer := make([]byte, int(chunkSize))
		r.file.Read(chunkBuffer)

		hash := sha256.Sum256(chunkBuffer)
		f.Write(hash[:])

	}

	return nil
}
