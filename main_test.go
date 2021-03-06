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
	v1, err := io.Open("mock.txt")
	if err != nil {
		t.Fatal("Expected to be able to read the original file")
	}

	// Fill signature in memory
	sig := sync.BuildSigTable(v1)
	// Write signatures
	// Simulation step for signatures write and read
	// Simulate split operation for signatures
	IO.WriteSignature("signature.bin", sig)

	// Sometime later :)
	// Expected receive same signature from original written file
	signaturesFromFile, _ := IO.ReadSignature("signature.bin")
	if !reflect.DeepEqual(sig, signaturesFromFile) {
		t.Errorf("Expected written signatures equal to out signatures")
	}

	v2, err := io.Open("mockV2.txt")
	if err != nil {
		t.Fatal("Expected to be able to read the V2 file")
	}

	// Match in block 2 the change "added"
	// V1 "i am here guys how are you doing this is a small test for chunk split and rolling hash"
	// V2 "i am here guys how are you doingadded this is a small test for chunk split and rolling hash"
	delta := sync.Delta(signaturesFromFile, v2)
	if string(delta[2].Lit) != "added" {
		t.Fatal("Expected match change from original in V2 file")
	}

}

func BenchmarkDelta(b *testing.B) {
	b.StopTimer()       // We are not analyzing io/declarations/etc
	blockSize := 1 << 4 // 16 bytes
	io := IO.New(blockSize)
	sync := Sync.New(blockSize)

	// Memory performance improvement using bufio.Reader
	v1, err := io.Open("mock.txt")
	if err != nil {
		panic("Fail opening mock.txt")
	}

	v2, err := io.Open("mockV2.txt")
	if err != nil {
		panic("Fail opening mockV2.txt")
	}

	b.StartTimer() // Start timer here to evaluate delta
	for i := 0; i <= b.N; i++ {
		sig := sync.BuildSigTable(v1)
		sync.Delta(sig, v2)
	}

}
