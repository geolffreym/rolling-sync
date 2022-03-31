package main

import (
	"rolling/io"
)

func main() {
	// Read file to split in chunks
	const bufferSize = (1 << 6) // chunk size = 64
	rolling := &io.IO{
		FileDir:       "test.txt",
		SignatureFile: "signature.bin",
		BufferSize:    bufferSize,
	}

	r, err := rolling.Read()
	if err != nil {
		panic(err)
	}

	r.WriteSignature()

}
