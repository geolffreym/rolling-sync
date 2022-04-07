/**
Copyright (c) 2022, Geolffrey Mena <gmjun2000@gmail.com>

Refs:
https://rsync.samba.org/tech_report/
https://www.zlib.net/maxino06_fletcher-adler.pdf
https://www.sciencedirect.com/science/article/pii/S1742287606000764#fig2
https://en.wikipedia.org/wiki/Adler-32
https://xilinx.github.io/Vitis_Libraries/security/2020.2/guide_L1/internals/adler32.html
**/
package main

import (
	"fmt"
	"log"
	IO "rolling/io"
	Sync "rolling/sync"
)

func main() {
	// Read file to split in chunks
	// chunk size = 64
	blockSize := 1 << 6
	io := IO.New(blockSize)
	sync := Sync.New(blockSize)

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

	// Roll it and compare the result with full re-calculus every time
	// For each block slice from file
	for _, block := range blocks {
		adler, md5 := sync.Signature(block)
		fmt.Printf("%d %s", adler, string(md5))
	}

}
