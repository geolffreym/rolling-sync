/**
Copyright (c) 2022, Geolffrey Mena <gmjun2000@gmail.com>

Refs:
https://rsync.samba.org/tech_report/
https://en.wikipedia.org/wiki/Adler-32
https://www.zlib.net/maxino06_fletcher-adler.pdf
https://www.sciencedirect.com/science/article/pii/S1742287606000764#fig2
https://xilinx.github.io/Vitis_Libraries/security/2020.2/guide_L1/internals/adler32.html
**/
package main

import (
	"log"
	IO "rolling/fileio"
	Sync "rolling/sync"
)

func main() {
	// Read file to split in chunks
	blockSize := 1 << 4 // 16 bytes
	io := IO.New(blockSize)
	sync := Sync.New(blockSize)

	// Memory performance improvement using bufio.Reader
	reader, err := io.Open("test.txt")
	if err != nil {
		log.Fatal(err)
	}

	// For each block slice from file
	sync.FillTable(reader)
	signatures := sync.Signatures()
	// checksums := make(map[uint32]map[string]int)
	// io.Signature.Write("signature.bin", )
	// fmt.Print(io.Signature.Read("signature.bin"))

	// End step 1

	newFile, err := io.Open("test2.txt")
	// out, err := io.Writer("delta.txt")
	sync.Delta(signatures, newFile)

}
