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
	IO "github.com/geolffreym/rolling-sync/fileio"
	Sync "github.com/geolffreym/rolling-sync/sync"
)

func main() {

	// Example usage
	blockSize := 1 << 4 // 16 bytes
	io := IO.New(blockSize)
	sync := Sync.New(blockSize)

	v1, err := io.Open("mock.txt")
	if err != nil {
		panic("Fail opening mock.txt")
	}

	v2, err := io.Open("mockV2.txt")
	if err != nil {
		panic("Fail opening mockV2.txt")
	}

	sig := sync.BuildSigTable(v1) // Signature file for "source"
	sync.Delta(sig, v2)           // Return delta with "sig" and "target" differences

}
