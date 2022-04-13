package main

import (
	"reflect"
	"testing"

	IO "github.com/geolffreym/rolling-sync/fileio"
	Sync "github.com/geolffreym/rolling-sync/sync"
)

func TestIntegration(t *testing.T) {
	blockSize := 1 << 4 // 16 bytes
	io := IO.New(blockSize)
	sync := Sync.New(blockSize)
	// Memory performance improvement using bufio.Reader
	original, err := io.Open("mock.txt")
	if err != nil {
		t.Fatal("Expected to be able to read the original file")
	}

	// Fill signature in memory
	sync.FillTable(original)
	// Retrieve signatures
	signatures := sync.Signatures()
	// Write signatures
	// Simulation step for signatures write and read
	// Simulate split operation for signatures
	IO.WriteSignature("signature.bin", signatures)

	// Sometime later :)
	// Expected receive same signature from original written file
	signaturesFromFile, _ := IO.ReadSignature("signature.bin")
	if !reflect.DeepEqual(signatures, signaturesFromFile) {
		t.Errorf("Expected written signatures equal to out signatures")
	}

	newFile, err := io.Open("mockV2.txt")
	if err != nil {
		t.Fatal("Expected to be able to read the V2 file")
	}

	delta := sync.Delta(signaturesFromFile, newFile)
	if string(delta[2].Lit) != "added" {
		t.Fatal("Expected match change from original in V2 file")
	}

}
