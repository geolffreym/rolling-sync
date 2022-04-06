package main

import (
	RollingIO "rolling/io"
	RollingHash "rolling/rolling"
)

func main() {
	// Read file to split in chunks
	// chunk size = 32
	io := &RollingIO.IO{BlockSize: (1 << 5)}
	rolling := RollingHash.New()

	// Memory performance improvement using bufio.Reader
	source, _ := io.Read("test.txt")
	target, _ := io.Read("test2.txt")
	offset := len(source)

	// Keep ratio for longest input
	if offset < len(target) {
		offset = len(target)
	}

	// Keep size ratio
	t := make([][]byte, 0, offset)
	copy(target, t)

	// Roll it and compare the result with full re-calculus every time
	// For each block slice from file
	for j, block := range source {
		// fmt.Printf("%d blocks processed: \n", j+1)
		// Initialize with small window chunk
		rolling.Compute(block, target[j])

	}
}
