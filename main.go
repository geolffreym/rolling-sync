/**
Copyright (c) 2022, Geolffrey Mena <gmjun2000@gmail.com>

Refs:
https://www.zlib.net/maxino06_fletcher-adler.pdf
https://www.sciencedirect.com/science/article/pii/S1742287606000764#fig2
https://en.wikipedia.org/wiki/Adler-32
https://xilinx.github.io/Vitis_Libraries/security/2020.2/guide_L1/internals/adler32.html
https://en.wikipedia.org/wiki/Rolling_hash#Cyclic_polynomial
**/
package main

import (
	"log"
	Adler32 "rolling/adler32"
	IO "rolling/io"
)

func main() {
	// Read file to split in chunks
	// chunk size = 32
	blockSize := 1 << 5
	io := IO.New(blockSize)
	rolling := Adler32.New()

	// Memory performance improvement using bufio.Reader
	blocks, err := io.Blocks("test.txt")
	if err != nil {
		log.Fatal(err)
	}

	// Performed writing operations
	writer, err := io.Writer("signature.txt")
	if err != nil {
		log.Fatal(err)
	}

	defer writer.Flush()

	n := 32
	// Roll it and compare the result with full re-calculus every time
	// For each block slice from file
	for _, block := range blocks {
		rolling.Reset()
		rolling.Write(block[:n])

		// for i := n; i < len(block); i++ {
		// 	rolling.Roll(block[i])
		// 	fmt.Printf("%s => %x\n", block[i-n+1:i+1], rolling.Sum())
		// }
	}

}
