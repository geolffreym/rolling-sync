package sync

import (
	"github.com/geolffreym/rolling-sync/fileio"
	"log"
	"testing"
)

/**
Test for matching differences

a mock.txt
b mockV2.txt
a= i am here guys how are you doing this is a small test for chunk split and rolling hash
b= i here guys how are you doing this is a small test chunk split and rolling hash

0 = i am here guys h
1 = ow are you doing
2 = this is a small
3 =  test for chunk
4 = split and rollin
5 = g hash

Signatures:
[
	0: {768804235 5d7b9b82d3dd8c4d13a576f004318130},
	1: {828311020 eb535617e82301559e56a18993cdbe39},
	2: {763037070 f38bd8f1d59e45f4a7bdaa6311064573},
	3: {800720288 6f2fcd27d23f5e98f486ff34ad580d09},
	4: {880805423 5f19d42bfb610b9861ec0704b6467910},
	5: {489488949 3024133c176e89ed84401db125a62ed0}
]


**/

func CalculateDelta(blockSize int, a string, b string) (map[int][]byte, error) {

	io := fileio.New(blockSize)
	sync := New(blockSize)

	// Memory performance improvement using bufio.Reader
	reader, err := io.Open(a)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// For each block slice from file
	sync.FillTable(reader)
	signatures := sync.Signatures()
	newFile, err := io.Open(b)
	return sync.Delta(signatures, newFile), nil
}

func TestDetectChunkChange(t *testing.T) {

	// // Read file to split in chunks
	// blockSize := 1 << 4 // 16 bytes
	// io := IO.New(blockSize)
	// sync := Sync{blockSize: blockSize}

	// // Memory performance improvement using bufio.Reader
	// reader, err := io.Open("../mock.txt")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // For each block slice from file
	// sync.FillTable(reader)
	// signatures := sync.Signatures()
	// newFile, err := io.Open("../mockV2.txt")
	// sync.Delta(signatures, newFile)

}

func TestDetectChunkRemoval(t *testing.T) {

}

func TestDetectChunkShifted(t *testing.T) {

}
