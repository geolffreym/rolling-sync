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
	"os"
	"runtime"
	"runtime/pprof"

	IO "github.com/geolffreym/rolling-sync/fileio"
	Sync "github.com/geolffreym/rolling-sync/sync"
)

func main() {

	// Performance test
	cpufile, err := os.Create("cpu.pprof")
	err = pprof.StartCPUProfile(cpufile)
	if err != nil {
		panic(err)
	}

	defer cpufile.Close()
	defer pprof.StopCPUProfile()

	blockSize := 1 << 4 // 16 bytes
	io := IO.New(blockSize)
	sync := Sync.New(blockSize)

	// Memory performance improvement using bufio.Reader
	reader, err := io.Open("mock.txt")
	newFile, err := io.Open("mockV2.txt")

	for i := 0; i <= 100000; i++ {

		// For each block slice from file
		sync.FillTable(reader)
		signatures := sync.Signatures()
		// TODO write delta if needed
		sync.Delta(signatures, newFile)
	}

	f, err := os.Create("mem.proof")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f.Close() // error handling omitted for example
	runtime.GC()    // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}

}
