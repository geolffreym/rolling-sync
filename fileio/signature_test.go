package fileio

import (
	"log"
	"reflect"
	Sync "rolling/sync"
	"testing"
)

func TestSignatureReadWrite(t *testing.T) {
	// Read file to split in chunks
	blockSize := 1 << 4 // 16 bytes
	io := IO{blockSize: 1 << 4}
	sync := Sync.New(blockSize)

	// Memory performance improvement using bufio.Reader
	reader, err := io.Open("mock.txt")
	if err != nil {
		log.Fatal(err)
	}

	// For each block slice from file
	sync.FillTable(reader)
	signatures := sync.Signatures()
	io.Signature.Write("signature.bin", signatures)
	out, err := io.Signature.Read("signature.bin")

	if !reflect.DeepEqual(signatures, out) {
		t.Errorf("Expected written signatures equal to out signatures")
	}

}
